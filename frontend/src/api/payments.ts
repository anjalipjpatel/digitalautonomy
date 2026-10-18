export type CreatePaymentRequest = {
  from_user: string
  to_user: string
  amount: string
  currency: 'GBP'
}

export type PaymentApiResponse = {
  consent?: {
    Data?: {
      ConsentId?: string
      Status?: string
    }
  }
  payment?: {
    Data?: {
      DomesticPaymentId?: string
      Status?: string
    }
  }
}

export async function createPayment(
  payment: CreatePaymentRequest,
): Promise<PaymentApiResponse> {
  const response = await fetch('/payments', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payment),
  })
  const responseBody = await response.text()

  if (!response.ok) {
    throw new Error(responseBody || `Payment API returned HTTP ${response.status}`)
  }

  try {
    return JSON.parse(responseBody) as PaymentApiResponse
  } catch {
    throw new Error('The payment API returned an invalid JSON response.')
  }
}
