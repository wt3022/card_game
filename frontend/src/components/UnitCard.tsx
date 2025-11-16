import type { Unit } from '../gen/common_pb'
import { Trait } from '../gen/common_pb'
import './UnitCard.css'

interface UnitCardProps {
  unit: Unit
  onClick: () => void
  isSelected: boolean
  isClickable: boolean
}

const traitLabels: Record<number, string> = {
  [Trait.RUSH]: '疾走',
  [Trait.DIRECT]: '直接',
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
  const canAttack = unit.attacksRemaining > 0 && (hasRush || !unit.summonedThisTurn)

  return (
    <div
      className={`unit-card ${isSelected ? 'selected' : ''} ${canAttack ? 'can-attack' : ''} ${
        isClickable ? 'clickable' : ''
      }`}
      onClick={isClickable ? onClick : undefined}
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

      {unit.effect && <div className="unit-effect">{unit.effect}</div>}

      <div className="unit-footer">
        <span className="attacks-remaining">
          攻撃: {unit.attacksRemaining}
        </span>
        {unit.summonedThisTurn && !hasRush && <span className="summoned-badge">召喚酔い</span>}
      </div>
    </div>
  )
}

