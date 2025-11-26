import type { Unit } from '../gen/common_pb'
import { Trait } from '../gen/common_pb'
import { extractEffectTimingsFromText } from '../utils/effectTimingUtils'
import { EffectTimingList } from './EffectTimingBadge'
import './UnitCard.css'

interface UnitCardProps {
  unit: Unit
  onClick: () => void
  isSelected: boolean
  isClickable: boolean
}

const traitLabels: Record<number, string> = {
  [Trait.RUSH]: '疾走',
  [Trait.CHARGE]: '突進',
  [Trait.WINDFURY]: '疾風',
  [Trait.PIERCE]: '貫通',
  [Trait.GUARDIAN]: '守護',
  [Trait.EFFECT_SHIELD]: '効果盾',
  [Trait.UNTARGETABLE]: '対象不可',
}

export default function UnitCard({
  unit,
  onClick,
  isSelected,
  isClickable,
}: UnitCardProps) {
  const hasRush = unit.traits.includes(Trait.RUSH)
  const hasCharge = unit.traits.includes(Trait.CHARGE)
  const canAttack =
    unit.attacksRemaining > 0 &&
    (hasRush || hasCharge || !unit.summonedThisTurn)

  // 効果テキストからタイミング情報を推測
  const effectTimings = unit.effect
    ? extractEffectTimingsFromText(unit.effect)
    : []

  // 場に出たユニットは召喚時効果（Immediate）を表示しない
  const fieldEffectTimings = effectTimings.filter((timing) => timing !== 1) // 1 = IMMEDIATE

  // 召喚時のみの効果テキストかどうか判定（トークン召喚など）
  const isOnlyImmediateEffect =
    unit.effect &&
    (unit.effect.includes('召喚') || unit.effect.includes('トークン')) &&
    effectTimings.length > 0 &&
    effectTimings.every((timing) => timing === 1)

  const interactiveProps = isClickable
    ? {
        role: 'button' as const,
        tabIndex: 0,
        onClick,
        onKeyDown: (e: React.KeyboardEvent) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            onClick()
          }
        },
      }
    : {}

  return (
    <div
      className={`unit-card ${isSelected ? 'selected' : ''} ${canAttack ? 'can-attack' : ''} ${
        isClickable ? 'clickable' : ''
      }`}
      {...interactiveProps}
    >
      <div className="unit-header">
        <div className="unit-name">{unit.name}</div>
        <div className="unit-cost">{unit.cost}</div>
      </div>

      <div className="unit-stats">
        <div className="stat attack">{unit.attack}</div>
        <div className="stat defense">
          {unit.currentDefense}/{unit.defense}
        </div>
      </div>

      {unit.traits.length > 0 && (
        <div className="unit-traits">
          {unit.traits.map((trait) => (
            <span key={trait} className="trait-badge">
              {traitLabels[trait] || trait}
            </span>
          ))}
        </div>
      )}

      {unit.effect && !isOnlyImmediateEffect && (
        <div className="unit-effect">
          {fieldEffectTimings.length > 0 && (
            <EffectTimingList
              timings={fieldEffectTimings}
              showLabel={false}
              showTooltip={true}
            />
          )}
          <span>{unit.effect}</span>
        </div>
      )}

      <div className="unit-footer">
        <span className="attacks-remaining">
          攻撃可能回数: {unit.attacksRemaining}
        </span>
        {unit.summonedThisTurn && !hasRush && !hasCharge && (
          <span className="summoned-badge">召喚酔い</span>
        )}
      </div>
    </div>
  )
}
