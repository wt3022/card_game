package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"card_game/internal/core/entity"
	"card_game/internal/core/port"
	"card_game/internal/infrastructure/persistence/model"

	"gorm.io/gorm"
)

// cardRepository カードリポジトリの実装
type cardRepository struct {
	db *gorm.DB
}

// NewCardRepository 新しいカードリポジトリを作成
func NewCardRepository(db *gorm.DB) port.CardRepository {
	return &cardRepository{db: db}
}

// toEntityModel CardModelをentity.Cardに変換
func toEntityCard(cardModel *model.CardModel) (*entity.Card, error) {
	card := &entity.Card{
		ID:      cardModel.ID,
		Name:    cardModel.Name,
		Type:    entity.CardType(cardModel.Type),
		Cost:    cardModel.Cost,
		Attack:  cardModel.Attack,
		Defense: cardModel.Defense,
		Effect:  cardModel.Effect,
	}

	// Traitsを変換
	traits := make([]entity.Trait, len(cardModel.Traits))
	for i, ct := range cardModel.Traits {
		traits[i] = entity.Trait(ct.Trait)
	}
	card.Traits = traits

	// CardEffectを変換（必要に応じて実装）
	// ...

	return card, nil
}

// toGormModel entity.CardをCardModelに変換
func toGormCard(card *entity.Card) *model.CardModel {
	cardModel := &model.CardModel{
		ID:      card.ID,
		Name:    card.Name,
		Type:    string(card.Type),
		Cost:    card.Cost,
		Attack:  card.Attack,
		Defense: card.Defense,
		Effect:  card.Effect,
	}

	// Traitsを変換
	traits := make([]model.CardTraitModel, len(card.Traits))
	for i, trait := range card.Traits {
		traits[i] = model.CardTraitModel{
			CardID: card.ID,
			Trait:  string(trait),
		}
	}
	cardModel.Traits = traits

	return cardModel
}

// attachCardEffect Cardエンティティに効果定義を付与
func (r *cardRepository) attachCardEffect(card *entity.Card) error {
	if card == nil {
		return nil
	}

	// CardEffectModelから説明を直接生成（常に最新の説明を取得）
	cardEffectModel, err := r.LoadCardEffectModel(card.ID)
	if err != nil {
		return err
	}

	if cardEffectModel != nil {
		// CardEffectModelから説明を生成してcard.Effectを上書き
		effectDescription := GenerateEffectDescription(cardEffectModel)
		if effectDescription != "" {
			card.Effect = effectDescription
		}
	}

	// entity.CardEffectも必要な場合はロード
	effect, err := r.loadCardEffect(card.ID)
	if err != nil {
		return err
	}
	card.CardEffect = effect

	return nil
}

// SaveCardEffect CardEffectを保存（model構造またはproto構造を受け取る）
func (r *cardRepository) SaveCardEffect(cardID string, cardEffect *entity.CardEffect) error {
	if cardEffect == nil || len(cardEffect.Definitions) == 0 {
		// CardEffectが存在しない場合は既存のものを削除
		if err := r.db.Where("card_id = ?", cardID).Delete(&model.CardEffectModel{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing card effect: %w", err)
		}
		return nil
	}

	// entity.CardEffectをmodel.CardEffectModelに変換
	cardEffectModel, err := CardEffectFromEntityToModel(cardEffect)
	if err != nil {
		return fmt.Errorf("failed to convert card effect to model: %w", err)
	}

	// 既存のCardEffectを削除
	if err := r.db.Where("card_id = ?", cardID).Delete(&model.CardEffectModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete existing card effect: %w", err)
	}

	// CardEffectModelを作成
	cardEffectModel.CardID = cardID
	if err := r.db.Create(cardEffectModel).Error; err != nil {
		return fmt.Errorf("failed to create card effect: %w", err)
	}

	// EffectDefinitionを保存
	for i := range cardEffectModel.Definitions {
		if err := r.saveEffectDefinitionModel(cardEffectModel.ID, &cardEffectModel.Definitions[i]); err != nil {
			return fmt.Errorf("failed to save effect definition: %w", err)
		}
	}

	return nil
}

// saveEffectDefinitionModel EffectDefinitionModelを保存
func (r *cardRepository) saveEffectDefinitionModel(cardEffectID uint, def *model.EffectDefinitionModel) error {
	// 既存IDをリセットして新規作成扱いにする
	def.ID = 0
	// RootNodeを保存
	var rootNodeID *uint
	if def.Root != nil {
		nodeID, err := r.saveEffectChainNodeModel(def.Root)
		if err != nil {
			return fmt.Errorf("failed to save root node: %w", err)
		}
		rootNodeID = &nodeID
	}

	// EffectDefinitionModelを作成
	def.CardEffectID = cardEffectID
	def.RootNodeID = rootNodeID
	if err := r.db.Create(def).Error; err != nil {
		return fmt.Errorf("failed to create effect definition: %w", err)
	}

	return nil
}

// saveEffectChainNodeModel EffectChainNodeModelを保存（再帰的）
func (r *cardRepository) saveEffectChainNodeModel(node *model.EffectChainNodeModel) (uint, error) {
	// AtomicEffectを保存
	var atomicEffectID *uint
	if node.AtomicEffect != nil {
		aeID, err := r.saveAtomicEffectModel(node.AtomicEffect)
		if err != nil {
			return 0, fmt.Errorf("failed to save atomic effect: %w", err)
		}
		atomicEffectID = &aeID
	}

	// EffectChainNodeModelを作成
	nodeModel := &model.EffectChainNodeModel{
		Type:           node.Type,
		AtomicEffectID: atomicEffectID,
	}
	if err := r.db.Create(nodeModel).Error; err != nil {
		return 0, fmt.Errorf("failed to create effect chain node: %w", err)
	}

	// ノードタイプに応じた具体テーブルに保存
	switch node.Type {
	case "THEN":
		var nextID *uint
		if node.Sequential != nil && node.Sequential.Next != nil {
			nextNodeID, err := r.saveEffectChainNodeModel(node.Sequential.Next)
			if err != nil {
				return 0, err
			}
			nextID = &nextNodeID
		}
		// THENノードは必ずsequential_nodesテーブルにレコードを作成
		seqModel := &model.SequentialNodeModel{
			NodeID: nodeModel.ID,
			NextID: nextID,
		}
		if err := r.db.Create(seqModel).Error; err != nil {
			return 0, fmt.Errorf("failed to create sequential node: %w", err)
		}
	case "AND":
		// Childrenを保存
		var childIDs []uint
		if node.Parallel != nil {
			for _, child := range node.Parallel.Children {
				if child != nil {
					childID, err := r.saveEffectChainNodeModel(child)
					if err != nil {
						return 0, fmt.Errorf("failed to save parallel child node: %w", err)
					}
					childIDs = append(childIDs, childID)
				}
			}
		}

		// ParallelNextを保存
		var parallelNextID *uint
		if node.Parallel != nil && node.Parallel.ParallelNext != nil {
			nextNodeID, err := r.saveEffectChainNodeModel(node.Parallel.ParallelNext)
			if err != nil {
				return 0, fmt.Errorf("failed to save parallel next node: %w", err)
			}
			parallelNextID = &nextNodeID
		}

		// ANDノードは必ずparallel_nodesテーブルにレコードを作成
		parModel := &model.ParallelNodeModel{
			NodeID:         nodeModel.ID,
			ParallelNextID: parallelNextID,
		}
		if err := r.db.Create(parModel).Error; err != nil {
			return 0, fmt.Errorf("failed to create parallel node: %w", err)
		}

		// 中間テーブルにChildrenを保存
		for _, childID := range childIDs {
			childModel := &model.ParallelNodeChildModel{
				ParallelNodeID: nodeModel.ID,
				ChildNodeID:    childID,
			}
			if err := r.db.Create(childModel).Error; err != nil {
				return 0, fmt.Errorf("failed to create parallel node child: %w", err)
			}
		}
	case "IF_ELSE":
		if node.IfElse != nil {
			if node.IfElse.Condition == nil {
				return 0, fmt.Errorf("if_else node requires condition")
			}
			if node.IfElse.Then == nil {
				return 0, fmt.Errorf("if_else node requires then node")
			}

			// Conditionを保存
			conditionID, err := r.saveConditionModel(node.IfElse.Condition)
			if err != nil {
				return 0, fmt.Errorf("failed to save if_else condition: %w", err)
			}

			// Thenを保存
			thenID, err := r.saveEffectChainNodeModel(node.IfElse.Then)
			if err != nil {
				return 0, fmt.Errorf("failed to save if_else then node: %w", err)
			}

			// Elseを保存（省略可能）
			var elseID *uint
			if node.IfElse.Else != nil {
				eID, err := r.saveEffectChainNodeModel(node.IfElse.Else)
				if err != nil {
					return 0, fmt.Errorf("failed to save if_else else node: %w", err)
				}
				elseID = &eID
			}

			ifElseModel := &model.IfElseNodeModel{
				NodeID:      nodeModel.ID,
				ThenID:      thenID,
				ElseID:      elseID,
				ConditionID: conditionID,
			}
			if err := r.db.Create(ifElseModel).Error; err != nil {
				return 0, fmt.Errorf("failed to create if_else node: %w", err)
			}
		}
	case "REPEAT":
		if node.Repeat != nil {
			if node.Repeat.RepeatEffect == nil {
				return 0, fmt.Errorf("repeat node requires repeat effect")
			}
			if node.Repeat.RepeatCount <= 0 {
				return 0, fmt.Errorf("repeat node requires positive count")
			}

			// RepeatEffectを保存
			repeatEffectID, err := r.saveEffectChainNodeModel(node.Repeat.RepeatEffect)
			if err != nil {
				return 0, fmt.Errorf("failed to save repeat effect node: %w", err)
			}

			repeatModel := &model.RepeatNodeModel{
				NodeID:         nodeModel.ID,
				RepeatEffectID: repeatEffectID,
				RepeatCount:    node.Repeat.RepeatCount,
			}
			if err := r.db.Create(repeatModel).Error; err != nil {
				return 0, fmt.Errorf("failed to create repeat node: %w", err)
			}
		}
	case "FOREACH":
		if node.ForEach != nil {
			if node.ForEach.ForEachEffect == nil {
				return 0, fmt.Errorf("for_each node requires for_each effect")
			}
			if node.ForEach.ForEachTarget == nil {
				return 0, fmt.Errorf("for_each node requires for_each target")
			}

			// ForEachEffectを保存
			forEachEffectID, err := r.saveEffectChainNodeModel(node.ForEach.ForEachEffect)
			if err != nil {
				return 0, fmt.Errorf("failed to save for_each effect node: %w", err)
			}

			// ForEachTargetを保存
			forEachTargetID, err := r.saveTargetSelectorModel(node.ForEach.ForEachTarget)
			if err != nil {
				return 0, fmt.Errorf("failed to save for_each target: %w", err)
			}

			forEachModel := &model.ForEachNodeModel{
				NodeID:          nodeModel.ID,
				ForEachEffectID: forEachEffectID,
				ForEachTargetID: forEachTargetID,
			}
			if err := r.db.Create(forEachModel).Error; err != nil {
				return 0, fmt.Errorf("failed to create for_each node: %w", err)
			}
		}
	}

	return nodeModel.ID, nil
}

// saveAtomicEffectModel AtomicEffectModelを保存
func (r *cardRepository) saveAtomicEffectModel(effect *model.AtomicEffectModel) (uint, error) {
	// TargetSelectorを保存
	targetID, err := r.saveTargetSelectorModel(effect.Target)
	if err != nil {
		return 0, fmt.Errorf("failed to save target selector: %w", err)
	}

	// Conditionを保存
	var conditionID *uint
	if effect.Condition != nil {
		condID, err := r.saveConditionModel(effect.Condition)
		if err != nil {
			return 0, fmt.Errorf("failed to save condition: %w", err)
		}
		conditionID = &condID
	}

	// Parametersを処理（空文字列、null、または無効なJSONの場合は空のJSONオブジェクトに変換）
	parameters := effect.Parameters
	if parameters == "" || parameters == "null" || parameters == "NULL" {
		parameters = "{}" // MySQLのJSONカラムには空のJSONオブジェクトを設定
	}
	// 有効なJSONかどうかを簡単にチェック（空文字列でない場合）
	if parameters != "" && parameters != "{}" {
		// JSONの妥当性をチェック（簡易版）
		if !strings.HasPrefix(parameters, "{") && !strings.HasPrefix(parameters, "[") {
			parameters = "{}"
		}
	}

	// AtomicEffectModelを作成
	aeModel := &model.AtomicEffectModel{
		Type:        effect.Type,
		Value:       effect.Value,
		Multiplier:  effect.Multiplier,
		Duration:    effect.Duration,
		Timing:      effect.Timing,
		TargetID:    targetID,
		ConditionID: conditionID,
		Parameters:  parameters,
	}
	if err := r.db.Create(aeModel).Error; err != nil {
		return 0, fmt.Errorf("failed to create atomic effect: %w", err)
	}

	return aeModel.ID, nil
}

// saveTargetSelectorModel TargetSelectorModelを保存
func (r *cardRepository) saveTargetSelectorModel(selector *model.TargetSelectorModel) (uint, error) {
	// TargetFilterを保存
	var filterID *uint
	if selector.Filter != nil {
		filterIDVal, err := r.saveTargetFilterModel(selector.Filter)
		if err != nil {
			return 0, fmt.Errorf("failed to save target filter: %w", err)
		}
		filterID = &filterIDVal
	}

	// TargetSelectorModelを作成
	tsModel := &model.TargetSelectorModel{
		Type:        selector.Type,
		Count:       selector.Count,
		Random:      selector.Random,
		SelectByMax: selector.SelectByMax,
		SelectByMin: selector.SelectByMin,
		FilterID:    filterID,
	}
	if err := r.db.Create(tsModel).Error; err != nil {
		return 0, fmt.Errorf("failed to create target selector: %w", err)
	}

	return tsModel.ID, nil
}

// saveTargetFilterModel TargetFilterModelを保存
func (r *cardRepository) saveTargetFilterModel(filter *model.TargetFilterModel) (uint, error) {
	// TargetFilterModelを作成
	tfModel := &model.TargetFilterModel{
		MinAttack:  filter.MinAttack,
		MaxAttack:  filter.MaxAttack,
		MinDefense: filter.MinDefense,
		MaxDefense: filter.MaxDefense,
		MinCost:    filter.MinCost,
		MaxCost:    filter.MaxCost,
		CardType:   filter.CardType,
	}
	if err := r.db.Create(tfModel).Error; err != nil {
		return 0, fmt.Errorf("failed to create target filter: %w", err)
	}

	// HasTraitsを保存
	for _, traitModel := range filter.HasTraits {
		traitModel.FilterID = tfModel.ID
		traitModel.IsHasTrait = true
		if err := r.db.Create(&traitModel).Error; err != nil {
			return 0, fmt.Errorf("failed to create has trait: %w", err)
		}
	}

	// LackTraitsを保存
	for _, traitModel := range filter.LackTraits {
		traitModel.FilterID = tfModel.ID
		traitModel.IsHasTrait = false
		if err := r.db.Create(&traitModel).Error; err != nil {
			return 0, fmt.Errorf("failed to create lack trait: %w", err)
		}
	}

	return tfModel.ID, nil
}

// saveConditionModel ConditionModelを保存
func (r *cardRepository) saveConditionModel(condition *model.ConditionModel) (uint, error) {
	condModel := &model.ConditionModel{
		Type:     condition.Type,
		Operator: condition.Operator,
		Value:    condition.Value,
	}
	if err := r.db.Create(condModel).Error; err != nil {
		return 0, fmt.Errorf("failed to create condition: %w", err)
	}

	return condModel.ID, nil
}

// Create カードを作成
func (r *cardRepository) Create(card *entity.Card) error {
	cardModel := toGormCard(card)
	if err := r.db.Create(cardModel).Error; err != nil {
		return fmt.Errorf("failed to create card: %w", err)
	}

	// CardEffectはmodel構造で直接保存されるため、ここでは処理しない
	// ハンドラーでmodel構造を直接受け取って保存する

	return nil
}

// FindByID IDでカードを検索
func (r *cardRepository) FindByID(id string) (*entity.Card, error) {
	var cardModel model.CardModel
	if err := r.db.Preload("Traits").First(&cardModel, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("card not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find card by id: %w", err)
	}
	card, err := toEntityCard(&cardModel)
	if err != nil {
		return nil, err
	}
	if err := r.attachCardEffect(card); err != nil {
		return nil, err
	}
	return card, nil
}

// FindAll すべてのカードを取得
func (r *cardRepository) FindAll() ([]*entity.Card, error) {
	var cardModels []model.CardModel
	if err := r.db.Preload("Traits").Find(&cardModels).Error; err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	cards := make([]*entity.Card, len(cardModels))
	for i, cardModel := range cardModels {
		card, err := toEntityCard(&cardModel)
		if err != nil {
			return nil, fmt.Errorf("failed to convert card model to entity: %w", err)
		}
		if err := r.attachCardEffect(card); err != nil {
			return nil, err
		}
		cards[i] = card
	}
	return cards, nil
}

// FindByType タイプでカードを検索
func (r *cardRepository) FindByType(cardType entity.CardType) ([]*entity.Card, error) {
	var cardModels []model.CardModel
	if err := r.db.Preload("Traits").Where("type = ?", string(cardType)).Find(&cardModels).Error; err != nil {
		return nil, fmt.Errorf("failed to query cards by type: %w", err)
	}
	cards := make([]*entity.Card, len(cardModels))
	for i, cardModel := range cardModels {
		card, err := toEntityCard(&cardModel)
		if err != nil {
			return nil, fmt.Errorf("failed to convert card model to entity: %w", err)
		}
		if err := r.attachCardEffect(card); err != nil {
			return nil, err
		}
		cards[i] = card
	}
	return cards, nil
}

// Update カードを更新
func (r *cardRepository) Update(card *entity.Card) error {
	// 既存のTraitsを削除
	if err := r.db.Where("card_id = ?", card.ID).Delete(&model.CardTraitModel{}).Error; err != nil {
		return fmt.Errorf("failed to delete existing traits: %w", err)
	}

	// カードの基本情報を更新
	cardModel := &model.CardModel{
		ID:      card.ID,
		Name:    card.Name,
		Type:    string(card.Type),
		Cost:    card.Cost,
		Attack:  card.Attack,
		Defense: card.Defense,
		Effect:  card.Effect,
	}

	if err := r.db.Model(&model.CardModel{}).Where("id = ?", card.ID).Updates(cardModel).Error; err != nil {
		return fmt.Errorf("failed to update card: %w", err)
	}

	// 新しいTraitsを追加
	if len(card.Traits) > 0 {
		traits := make([]model.CardTraitModel, len(card.Traits))
		for i, trait := range card.Traits {
			traits[i] = model.CardTraitModel{
				CardID: card.ID,
				Trait:  string(trait),
			}
		}
		if err := r.db.Create(&traits).Error; err != nil {
			return fmt.Errorf("failed to create traits: %w", err)
		}
	}

	// CardEffectはmodel構造で直接保存されるため、ここでは処理しない
	// ハンドラーでmodel構造を直接受け取って保存する

	return nil
}

// Delete カードを削除
func (r *cardRepository) Delete(id string) error {
	if err := r.db.Delete(&model.CardModel{}, "id = ?", id).Error; err != nil {
		return fmt.Errorf("failed to delete card: %w", err)
	}
	return nil
}

// loadCardEffect 指定カードの効果定義を読み込む
func (r *cardRepository) loadCardEffect(cardID string) (*entity.CardEffect, error) {
	var cardEffect model.CardEffectModel
	if err := r.db.Where("card_id = ?", cardID).First(&cardEffect).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load card effect: %w", err)
	}

	definitions, err := r.loadEffectDefinitions(cardEffect.ID)
	if err != nil {
		return nil, err
	}
	if len(definitions) == 0 {
		return nil, nil
	}

	cardEffectEntity := &entity.CardEffect{
		Definitions: definitions,
	}
	cardEffectEntity.Description = cardEffectEntity.GenerateDescription()
	return cardEffectEntity, nil
}

func (r *cardRepository) loadEffectDefinitions(cardEffectID uint) ([]*entity.EffectDefinition, error) {
	var definitionModels []model.EffectDefinitionModel
	if err := r.db.Where("card_effect_id = ?", cardEffectID).Find(&definitionModels).Error; err != nil {
		return nil, fmt.Errorf("failed to load effect definitions: %w", err)
	}

	definitions := make([]*entity.EffectDefinition, 0, len(definitionModels))
	for _, defModel := range definitionModels {
		def := &entity.EffectDefinition{
			ID:            strconv.FormatUint(uint64(defModel.ID), 10),
			RequireTarget: defModel.RequireTarget,
		}
		if defModel.RootNodeID != nil {
			rootNode, err := r.loadEffectChainNode(*defModel.RootNodeID)
			if err != nil {
				return nil, err
			}
			def.Root = rootNode
		}
		definitions = append(definitions, def)
	}

	return definitions, nil
}

func (r *cardRepository) loadEffectChainNode(nodeID uint) (*entity.EffectChainNode, error) {
	var nodeModel model.EffectChainNodeModel
	if err := r.db.First(&nodeModel, nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load effect chain node (id=%d): %w", nodeID, err)
	}

	var atomicEffect *entity.AtomicEffect
	if nodeModel.AtomicEffectID != nil {
		ae, err := r.loadAtomicEffect(*nodeModel.AtomicEffectID)
		if err != nil {
			return nil, err
		}
		atomicEffect = ae
	}

	node := &entity.EffectChainNode{
		Type: entity.EffectOperator(nodeModel.Type),
	}

	switch node.Type {
	case entity.OperatorSequential:
		seq, err := r.loadSequentialNode(nodeModel.ID, atomicEffect)
		if err != nil {
			return nil, err
		}
		node.Sequential = seq
	case entity.OperatorParallel:
		par, err := r.loadParallelNode(nodeModel.ID)
		if err != nil {
			return nil, err
		}
		node.Parallel = par
	case entity.OperatorIfElse:
		ifElse, err := r.loadIfElseNode(nodeModel.ID)
		if err != nil {
			return nil, err
		}
		node.IfElse = ifElse
	case entity.OperatorRepeat:
		repeat, err := r.loadRepeatNode(nodeModel.ID)
		if err != nil {
			return nil, err
		}
		node.Repeat = repeat
	case entity.OperatorForEach:
		forEach, err := r.loadForEachNode(nodeModel.ID)
		if err != nil {
			return nil, err
		}
		node.ForEach = forEach
	default:
		return nil, fmt.Errorf("unsupported effect chain node type: %s", nodeModel.Type)
	}

	return node, nil
}

func (r *cardRepository) loadSequentialNode(nodeID uint, atomicEffect *entity.AtomicEffect) (*entity.SequentialNode, error) {
	var seqModel model.SequentialNodeModel
	if err := r.db.First(&seqModel, "node_id = ?", nodeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &entity.SequentialNode{Effect: atomicEffect}, nil
		}
		return nil, fmt.Errorf("failed to load sequential node: %w", err)
	}

	seq := &entity.SequentialNode{
		Effect: atomicEffect,
	}
	if seqModel.NextID != nil {
		nextNode, err := r.loadEffectChainNode(*seqModel.NextID)
		if err != nil {
			return nil, err
		}
		seq.Next = nextNode
	}

	return seq, nil
}

func (r *cardRepository) loadParallelNode(nodeID uint) (*entity.ParallelNode, error) {
	var parModel model.ParallelNodeModel
	if err := r.db.First(&parModel, "node_id = ?", nodeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// レコードがない場合は空のParallelNodeを返す
			return &entity.ParallelNode{Children: []*entity.EffectChainNode{}}, nil
		}
		return nil, fmt.Errorf("failed to load parallel node: %w", err)
	}

	parallel := &entity.ParallelNode{}

	var childRelations []model.ParallelNodeChildModel
	if err := r.db.Where("parallel_node_id = ?", nodeID).Find(&childRelations).Error; err != nil {
		return nil, fmt.Errorf("failed to load parallel node children: %w", err)
	}

	parallel.Children = make([]*entity.EffectChainNode, 0, len(childRelations))
	for _, childRel := range childRelations {
		childNode, err := r.loadEffectChainNode(childRel.ChildNodeID)
		if err != nil {
			return nil, err
		}
		parallel.Children = append(parallel.Children, childNode)
	}

	if parModel.ParallelNextID != nil {
		nextNode, err := r.loadEffectChainNode(*parModel.ParallelNextID)
		if err != nil {
			return nil, err
		}
		parallel.Next = nextNode
	}

	return parallel, nil
}

func (r *cardRepository) loadIfElseNode(nodeID uint) (*entity.IfElseNode, error) {
	var ifElseModel model.IfElseNodeModel
	if err := r.db.First(&ifElseModel, "node_id = ?", nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load if_else node: %w", err)
	}

	condition, err := r.loadCondition(ifElseModel.ConditionID)
	if err != nil {
		return nil, err
	}

	thenNode, err := r.loadEffectChainNode(ifElseModel.ThenID)
	if err != nil {
		return nil, err
	}

	var elseNode *entity.EffectChainNode
	if ifElseModel.ElseID != nil {
		elseNode, err = r.loadEffectChainNode(*ifElseModel.ElseID)
		if err != nil {
			return nil, err
		}
	}

	return &entity.IfElseNode{
		Condition: condition,
		Then:      thenNode,
		Else:      elseNode,
	}, nil
}

func (r *cardRepository) loadRepeatNode(nodeID uint) (*entity.RepeatNode, error) {
	var repeatModel model.RepeatNodeModel
	if err := r.db.First(&repeatModel, "node_id = ?", nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load repeat node: %w", err)
	}

	repeatEffect, err := r.loadEffectChainNode(repeatModel.RepeatEffectID)
	if err != nil {
		return nil, err
	}

	return &entity.RepeatNode{
		Count:  repeatModel.RepeatCount,
		Effect: repeatEffect,
	}, nil
}

func (r *cardRepository) loadForEachNode(nodeID uint) (*entity.ForEachNode, error) {
	var forEachModel model.ForEachNodeModel
	if err := r.db.First(&forEachModel, "node_id = ?", nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load for_each node: %w", err)
	}

	target, err := r.loadTargetSelector(forEachModel.ForEachTargetID)
	if err != nil {
		return nil, err
	}

	effectNode, err := r.loadEffectChainNode(forEachModel.ForEachEffectID)
	if err != nil {
		return nil, err
	}

	return &entity.ForEachNode{
		Target: *target,
		Effect: effectNode,
	}, nil
}

func (r *cardRepository) loadAtomicEffect(id uint) (*entity.AtomicEffect, error) {
	var atomicModel model.AtomicEffectModel
	if err := r.db.First(&atomicModel, id).Error; err != nil {
		return nil, fmt.Errorf("failed to load atomic effect: %w", err)
	}

	target, err := r.loadTargetSelector(atomicModel.TargetID)
	if err != nil {
		return nil, err
	}

	var condition *entity.Condition
	if atomicModel.ConditionID != nil {
		condition, err = r.loadCondition(*atomicModel.ConditionID)
		if err != nil {
			return nil, err
		}
	}

	parameters := make(map[string]any)
	if atomicModel.Parameters != "" {
		if err := json.Unmarshal([]byte(atomicModel.Parameters), &parameters); err != nil {
			parameters = make(map[string]any)
		}
	}

	return &entity.AtomicEffect{
		ID:         strconv.FormatUint(uint64(atomicModel.ID), 10),
		Type:       entity.AtomicEffectType(atomicModel.Type),
		Target:     *target,
		Value:      atomicModel.Value,
		Multiplier: atomicModel.Multiplier,
		Duration:   atomicModel.Duration,
		Timing:     entity.EffectTiming(atomicModel.Timing),
		Condition:  condition,
		Parameters: parameters,
	}, nil
}

func (r *cardRepository) loadTargetSelector(id uint) (*entity.TargetSelector, error) {
	var selector model.TargetSelectorModel
	if err := r.db.First(&selector, id).Error; err != nil {
		return nil, fmt.Errorf("failed to load target selector: %w", err)
	}

	var filter *entity.TargetFilter
	if selector.FilterID != nil {
		f, err := r.loadTargetFilter(*selector.FilterID)
		if err != nil {
			return nil, err
		}
		filter = f
	}

	return &entity.TargetSelector{
		Type:        entity.EffectTarget(selector.Type),
		Filter:      filter,
		Count:       selector.Count,
		Random:      selector.Random,
		SelectByMax: selector.SelectByMax,
		SelectByMin: selector.SelectByMin,
	}, nil
}

func (r *cardRepository) loadTargetFilter(id uint) (*entity.TargetFilter, error) {
	var filterModel model.TargetFilterModel
	if err := r.db.First(&filterModel, id).Error; err != nil {
		return nil, fmt.Errorf("failed to load target filter: %w", err)
	}

	filter := &entity.TargetFilter{
		MinAttack:  filterModel.MinAttack,
		MaxAttack:  filterModel.MaxAttack,
		MinDefense: filterModel.MinDefense,
		MaxDefense: filterModel.MaxDefense,
		MinCost:    filterModel.MinCost,
		MaxCost:    filterModel.MaxCost,
	}

	if filterModel.CardType != "" {
		cardType := entity.CardType(filterModel.CardType)
		filter.CardType = &cardType
	}

	var traitModels []model.TargetFilterTraitModel
	if err := r.db.Where("filter_id = ?", id).Find(&traitModels).Error; err != nil {
		return nil, fmt.Errorf("failed to load target filter traits: %w", err)
	}

	for _, traitModel := range traitModels {
		trait := entity.Trait(traitModel.Trait)
		if traitModel.IsHasTrait {
			filter.HasTrait = append(filter.HasTrait, trait)
		} else {
			filter.LackTrait = append(filter.LackTrait, trait)
		}
	}

	return filter, nil
}

func (r *cardRepository) loadCondition(id uint) (*entity.Condition, error) {
	var conditionModel model.ConditionModel
	if err := r.db.First(&conditionModel, id).Error; err != nil {
		return nil, fmt.Errorf("failed to load condition: %w", err)
	}

	return &entity.Condition{
		Type:     entity.ConditionType(conditionModel.Type),
		Operator: entity.ComparisonOperator(conditionModel.Operator),
		Value:    conditionModel.Value,
	}, nil
}

// LoadCardEffectModel CardEffectModelをロード（Protoへの変換用）
func (r *cardRepository) LoadCardEffectModel(cardID string) (*model.CardEffectModel, error) {
	var cardEffectModel model.CardEffectModel
	if err := r.db.Where("card_id = ?", cardID).First(&cardEffectModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to load card effect model: %w", err)
	}

	// Definitionsをロード
	definitions, err := r.loadEffectDefinitionModels(cardEffectModel.ID)
	if err != nil {
		return nil, err
	}
	cardEffectModel.Definitions = definitions

	return &cardEffectModel, nil
}

// loadEffectDefinitionModels EffectDefinitionModelsをロード
func (r *cardRepository) loadEffectDefinitionModels(cardEffectID uint) ([]model.EffectDefinitionModel, error) {
	var definitionModels []model.EffectDefinitionModel
	if err := r.db.Where("card_effect_id = ?", cardEffectID).Find(&definitionModels).Error; err != nil {
		return nil, fmt.Errorf("failed to load effect definition models: %w", err)
	}

	for i := range definitionModels {
		if definitionModels[i].RootNodeID != nil {
			rootNode, err := r.loadEffectChainNodeModel(*definitionModels[i].RootNodeID)
			if err != nil {
				return nil, err
			}
			definitionModels[i].Root = rootNode
		}
	}

	return definitionModels, nil
}

// loadEffectChainNodeModel EffectChainNodeModelをロード
func (r *cardRepository) loadEffectChainNodeModel(nodeID uint) (*model.EffectChainNodeModel, error) {
	var nodeModel model.EffectChainNodeModel
	if err := r.db.First(&nodeModel, nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load effect chain node model: %w", err)
	}

	// AtomicEffectをロード
	if nodeModel.AtomicEffectID != nil {
		atomicEffect, err := r.loadAtomicEffectModel(*nodeModel.AtomicEffectID)
		if err != nil {
			return nil, err
		}
		nodeModel.AtomicEffect = atomicEffect
	}

	// ノードタイプに応じた具体構造をロード
	switch nodeModel.Type {
	case "THEN":
		seq, err := r.loadSequentialNodeModel(nodeModel.ID)
		if err != nil {
			return nil, err
		}
		nodeModel.Sequential = seq
	case "AND":
		par, err := r.loadParallelNodeModel(nodeModel.ID)
		if err != nil {
			return nil, err
		}
		nodeModel.Parallel = par
	case "IF_ELSE":
		ifElse, err := r.loadIfElseNodeModel(nodeModel.ID)
		if err != nil {
			return nil, err
		}
		nodeModel.IfElse = ifElse
	case "REPEAT":
		repeat, err := r.loadRepeatNodeModel(nodeModel.ID)
		if err != nil {
			return nil, err
		}
		nodeModel.Repeat = repeat
	case "FOREACH":
		forEach, err := r.loadForEachNodeModel(nodeModel.ID)
		if err != nil {
			return nil, err
		}
		nodeModel.ForEach = forEach
	}

	return &nodeModel, nil
}

// loadSequentialNodeModel SequentialNodeModelをロード
func (r *cardRepository) loadSequentialNodeModel(nodeID uint) (*model.SequentialNodeModel, error) {
	var seqModel model.SequentialNodeModel
	if err := r.db.First(&seqModel, "node_id = ?", nodeID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &model.SequentialNodeModel{NodeID: nodeID}, nil
		}
		return nil, fmt.Errorf("failed to load sequential node model: %w", err)
	}

	if seqModel.NextID != nil {
		nextNode, err := r.loadEffectChainNodeModel(*seqModel.NextID)
		if err != nil {
			return nil, err
		}
		seqModel.Next = nextNode
	}

	return &seqModel, nil
}

// loadParallelNodeModel ParallelNodeModelをロード
func (r *cardRepository) loadParallelNodeModel(nodeID uint) (*model.ParallelNodeModel, error) {
	var parModel model.ParallelNodeModel
	if err := r.db.First(&parModel, "node_id = ?", nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load parallel node model: %w", err)
	}

	// 子ノードをロード
	var childRelations []model.ParallelNodeChildModel
	if err := r.db.Where("parallel_node_id = ?", nodeID).Find(&childRelations).Error; err != nil {
		return nil, fmt.Errorf("failed to load parallel node children: %w", err)
	}

	parModel.Children = make([]*model.EffectChainNodeModel, 0, len(childRelations))
	for _, childRel := range childRelations {
		childNode, err := r.loadEffectChainNodeModel(childRel.ChildNodeID)
		if err != nil {
			return nil, err
		}
		parModel.Children = append(parModel.Children, childNode)
	}

	if parModel.ParallelNextID != nil {
		nextNode, err := r.loadEffectChainNodeModel(*parModel.ParallelNextID)
		if err != nil {
			return nil, err
		}
		parModel.ParallelNext = nextNode
	}

	return &parModel, nil
}

// loadIfElseNodeModel IfElseNodeModelをロード
func (r *cardRepository) loadIfElseNodeModel(nodeID uint) (*model.IfElseNodeModel, error) {
	var ifElseModel model.IfElseNodeModel
	if err := r.db.First(&ifElseModel, "node_id = ?", nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load if_else node model: %w", err)
	}

	// Conditionをロード
	condition, err := r.loadConditionModel(ifElseModel.ConditionID)
	if err != nil {
		return nil, err
	}
	ifElseModel.Condition = condition

	// Thenをロード
	thenNode, err := r.loadEffectChainNodeModel(ifElseModel.ThenID)
	if err != nil {
		return nil, err
	}
	ifElseModel.Then = thenNode

	// Elseをロード
	if ifElseModel.ElseID != nil {
		elseNode, err := r.loadEffectChainNodeModel(*ifElseModel.ElseID)
		if err != nil {
			return nil, err
		}
		ifElseModel.Else = elseNode
	}

	return &ifElseModel, nil
}

// loadRepeatNodeModel RepeatNodeModelをロード
func (r *cardRepository) loadRepeatNodeModel(nodeID uint) (*model.RepeatNodeModel, error) {
	var repeatModel model.RepeatNodeModel
	if err := r.db.First(&repeatModel, "node_id = ?", nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load repeat node model: %w", err)
	}

	repeatEffect, err := r.loadEffectChainNodeModel(repeatModel.RepeatEffectID)
	if err != nil {
		return nil, err
	}
	repeatModel.RepeatEffect = repeatEffect

	return &repeatModel, nil
}

// loadForEachNodeModel ForEachNodeModelをロード
func (r *cardRepository) loadForEachNodeModel(nodeID uint) (*model.ForEachNodeModel, error) {
	var forEachModel model.ForEachNodeModel
	if err := r.db.First(&forEachModel, "node_id = ?", nodeID).Error; err != nil {
		return nil, fmt.Errorf("failed to load for_each node model: %w", err)
	}

	// ForEachTargetをロード
	target, err := r.loadTargetSelectorModel(forEachModel.ForEachTargetID)
	if err != nil {
		return nil, err
	}
	forEachModel.ForEachTarget = target

	// ForEachEffectをロード
	effectNode, err := r.loadEffectChainNodeModel(forEachModel.ForEachEffectID)
	if err != nil {
		return nil, err
	}
	forEachModel.ForEachEffect = effectNode

	return &forEachModel, nil
}

// loadAtomicEffectModel AtomicEffectModelをロード
func (r *cardRepository) loadAtomicEffectModel(id uint) (*model.AtomicEffectModel, error) {
	var atomicModel model.AtomicEffectModel
	if err := r.db.First(&atomicModel, id).Error; err != nil {
		return nil, fmt.Errorf("failed to load atomic effect model: %w", err)
	}

	// Targetをロード
	target, err := r.loadTargetSelectorModel(atomicModel.TargetID)
	if err != nil {
		return nil, err
	}
	atomicModel.Target = target

	// Conditionをロード
	if atomicModel.ConditionID != nil {
		condition, err := r.loadConditionModel(*atomicModel.ConditionID)
		if err != nil {
			return nil, err
		}
		atomicModel.Condition = condition
	}

	return &atomicModel, nil
}

// loadTargetSelectorModel TargetSelectorModelをロード
func (r *cardRepository) loadTargetSelectorModel(id uint) (*model.TargetSelectorModel, error) {
	var selector model.TargetSelectorModel
	if err := r.db.First(&selector, id).Error; err != nil {
		return nil, fmt.Errorf("failed to load target selector model: %w", err)
	}

	// Filterをロード
	if selector.FilterID != nil {
		filter, err := r.loadTargetFilterModel(*selector.FilterID)
		if err != nil {
			return nil, err
		}
		selector.Filter = filter
	}

	return &selector, nil
}

// loadTargetFilterModel TargetFilterModelをロード
func (r *cardRepository) loadTargetFilterModel(id uint) (*model.TargetFilterModel, error) {
	var filterModel model.TargetFilterModel
	if err := r.db.First(&filterModel, id).Error; err != nil {
		return nil, fmt.Errorf("failed to load target filter model: %w", err)
	}

	// Traitsをロード
	var traitModels []model.TargetFilterTraitModel
	if err := r.db.Where("filter_id = ?", id).Find(&traitModels).Error; err != nil {
		return nil, fmt.Errorf("failed to load target filter traits: %w", err)
	}

	for _, traitModel := range traitModels {
		if traitModel.IsHasTrait {
			filterModel.HasTraits = append(filterModel.HasTraits, traitModel)
		} else {
			filterModel.LackTraits = append(filterModel.LackTraits, traitModel)
		}
	}

	return &filterModel, nil
}

// loadConditionModel ConditionModelをロード
func (r *cardRepository) loadConditionModel(id uint) (*model.ConditionModel, error) {
	var conditionModel model.ConditionModel
	if err := r.db.First(&conditionModel, id).Error; err != nil {
		return nil, fmt.Errorf("failed to load condition model: %w", err)
	}

	return &conditionModel, nil
}

// GetCardEffectAsProto CardEffectをProto形式で取得
func (r *cardRepository) GetCardEffect(cardID string) (*entity.CardEffect, error) {
	cardEffectModel, err := r.LoadCardEffectModel(cardID)
	if err != nil {
		return nil, err
	}
	if cardEffectModel == nil {
		return nil, nil
	}

	// Model→Entity変換
	cardEffect, err := CardEffectFromModelToEntity(cardEffectModel)
	if err != nil {
		return nil, fmt.Errorf("failed to convert card effect to entity: %w", err)
	}

	return cardEffect, nil
}

// GenerateEffectDescription CardEffectから効果テキストを生成
func (r *cardRepository) GenerateEffectDescription(cardID string) (string, error) {
	cardEffectModel, err := r.LoadCardEffectModel(cardID)
	if err != nil {
		return "", err
	}
	if cardEffectModel == nil {
		return "", nil
	}

	description := GenerateEffectDescription(cardEffectModel)
	return description, nil
}
