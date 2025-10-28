/**
 * Formats a phone number to (XXX) XXX-XXXX format
 * @param value - The input phone number string
 * @returns Formatted phone number
 */
export function formatPhoneNumber(value: string): string {
  // Remove all non-numeric characters
  const phoneNumber = value.replace(/\D/g, '')

  // Limit to 10 digits
  const truncated = phoneNumber.slice(0, 10)

  // Format based on length
  if (truncated.length === 0) {
    return ''
  } else if (truncated.length <= 3) {
    return `(${truncated}`
  } else if (truncated.length <= 6) {
    return `(${truncated.slice(0, 3)}) ${truncated.slice(3)}`
  } else {
    return `(${truncated.slice(0, 3)}) ${truncated.slice(3, 6)}-${truncated.slice(6)}`
  }
}

/**
 * Extracts only digits from a phone number for storage
 * @param formattedPhone - The formatted phone number string
 * @returns Phone number with only digits
 */
export function unformatPhoneNumber(formattedPhone: string): string {
  return formattedPhone.replace(/\D/g, '')
}
