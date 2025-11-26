import './LoadingSpinner.css'

interface LoadingSpinnerProps {
  message?: string
  size?: 'small' | 'medium' | 'large'
}

export default function LoadingSpinner({
  message = '読み込み中...',
  size = 'medium',
}: LoadingSpinnerProps) {
  return (
    <div className={`loading-spinner-container ${size}`}>
      <div className="loading-spinner" />
      {message && <p className="loading-message">{message}</p>}
    </div>
  )
}
