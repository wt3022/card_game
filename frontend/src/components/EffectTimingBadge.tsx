import type React from 'react'
import { useRef, useState } from 'react'
import {
  getEffectTimingDescription,
  getEffectTimingIcon,
  getEffectTimingLabel,
} from '../constants/effectTiming'
import { EffectTiming } from '../gen/common_pb'
import './EffectTimingBadge.css'

interface EffectTimingBadgeProps {
  timing: EffectTiming
  showLabel?: boolean
  showTooltip?: boolean
}

/**
 * 効果発動タイミングを表示するバッジコンポーネント
 */
export function EffectTimingBadge({
  timing,
  showLabel = true,
  showTooltip = true,
}: EffectTimingBadgeProps): React.ReactNode {
  const [tooltipStyle, setTooltipStyle] = useState<React.CSSProperties>({
    position: 'fixed',
    left: '0px',
    top: '0px',
  })
  const badgeRef = useRef<HTMLSpanElement>(null)

  if (timing === EffectTiming.UNSPECIFIED) {
    return null
  }

  const icon = getEffectTimingIcon(timing)
  const label = getEffectTimingLabel(timing)
  const description = getEffectTimingDescription(timing)

  const className = `effect-timing-badge ${getTimingClassName(timing)}`

  const handleMouseEnter = () => {
    if (badgeRef.current && showTooltip) {
      const rect = badgeRef.current.getBoundingClientRect()
      setTooltipStyle({
        position: 'fixed',
        left: `${rect.left + rect.width / 2}px`,
        top: `${rect.top - 8}px`,
      })
    }
  }

  const badge = (
    <span className={className}>
      <span className="icon">{icon}</span>
      {showLabel && <span className="label">{label}</span>}
    </span>
  )

  if (!showTooltip) {
    return badge
  }

  return (
    <span
      ref={badgeRef}
      className="effect-timing-tooltip"
      onMouseEnter={handleMouseEnter}
      role="button"
      tabIndex={0}
    >
      {badge}
      <span className="tooltip-content" style={tooltipStyle}>
        {description}
      </span>
    </span>
  )
}

/**
 * タイミングに応じたCSSクラス名を取得
 */
function getTimingClassName(timing: EffectTiming): string {
  switch (timing) {
    case EffectTiming.IMMEDIATE:
      return 'immediate'
    case EffectTiming.ON_SUMMON:
      return 'on-summon'
    case EffectTiming.ON_DESTROY:
      return 'on-destroy'
    case EffectTiming.ON_ATTACK:
      return 'on-attack'
    case EffectTiming.ON_DAMAGED:
      return 'on-damaged'
    case EffectTiming.TURN_START:
      return 'turn-start'
    case EffectTiming.TURN_END:
      return 'turn-end'
    default:
      return ''
  }
}

interface EffectTimingListProps {
  timings: EffectTiming[]
  showLabel?: boolean
  showTooltip?: boolean
}

/**
 * 複数の効果タイミングをリスト表示
 */
export function EffectTimingList({
  timings,
  showLabel = true,
  showTooltip = true,
}: EffectTimingListProps): React.ReactNode {
  // 重複を除去してソート
  const uniqueTimings = Array.from(new Set(timings))
    .filter((t) => t !== EffectTiming.UNSPECIFIED)
    .sort((a, b) => a - b)

  if (uniqueTimings.length === 0) {
    return null
  }

  return (
    <div className="effect-timing-list">
      {uniqueTimings.map((timing) => (
        <EffectTimingBadge
          key={timing}
          timing={timing}
          showLabel={showLabel}
          showTooltip={showTooltip}
        />
      ))}
    </div>
  )
}
