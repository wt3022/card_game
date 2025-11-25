/**
 * エラーメッセージを整形
 */
export const formatErrorMessage = (
  error: unknown,
  defaultMessage: string,
): string => {
  if (error instanceof Error) {
    return error.message
  }
  return defaultMessage
}

/**
 * 遅延実行
 */
export const delay = (ms: number): Promise<void> => {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

/**
 * デバウンス処理
 */
export const debounce = <T extends (...args: Parameters<T>) => void>(
  func: T,
  wait: number,
): ((...args: Parameters<T>) => void) => {
  let timeout: NodeJS.Timeout | null = null

  return (...args: Parameters<T>) => {
    if (timeout) {
      clearTimeout(timeout)
    }

    timeout = setTimeout(() => {
      func(...args)
    }, wait)
  }
}
