import { useState, type FormEvent } from 'react'
import { ArrowDown, ArrowRight, Check, CircleAlert, LoaderCircle, ShieldCheck } from 'lucide-react'
import { createPayment, type PaymentApiResponse } from '@/api/payments'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { PersonPicker, type DemoUser } from '@/components/payment/person-picker'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Separator } from '@/components/ui/separator'
import { formatAmount } from '@/lib/utils'

const users: DemoUser[] = [
  { id: 'alice', name: 'Alice Morgan', initials: 'AM', account: 'Everyday account ·· 2841' },
  { id: 'bob', name: 'Bob Chen', initials: 'BC', account: 'Current account ·· 7193' },
]

function App() {
  const [fromUser, setFromUser] = useState(users[0].id)
  const [toUser, setToUser] = useState(users[1].id)
  const [amount, setAmount] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const [result, setResult] = useState<PaymentApiResponse | null>(null)

  const sender = users.find((user) => user.id === fromUser) ?? users[0]
  const recipient = users.find((user) => user.id === toUser) ?? users[1]
  const formattedAmount = formatAmount(amount)

  async function submitPayment(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setError('')
    setResult(null)

    const numericAmount = Number(amount)
    if (
      fromUser === toUser ||
      !/^\d+(?:\.\d{1,2})?$/.test(amount) ||
      !Number.isFinite(numericAmount) ||
      numericAmount <= 0
    ) {
      setError('Choose two different people and enter an amount greater than zero with up to two decimal places.')
      return
    }

    setSubmitting(true)
    try {
      setResult(await createPayment({
          from_user: fromUser,
          to_user: toUser,
          amount,
          currency: 'GBP',
      }))
    } catch (submitError) {
      setError(submitError instanceof Error ? submitError.message : 'Could not complete the demo payment.')
    } finally {
      setSubmitting(false)
    }
  }

  function changeSender(userId: string) {
    setFromUser(userId)
    setResult(null)
    setError('')
    if (toUser === userId) {
      const otherUser = users.find((user) => user.id !== userId)
      if (otherUser) setToUser(otherUser.id)
    }
  }

  function changeRecipient(userId: string) {
    setToUser(userId)
    setResult(null)
    setError('')
  }

  return (
    <main className="mx-auto max-w-5xl space-y-6 px-4 py-8">
      <header>
        <div>
          <p className="text-sm text-muted-foreground">
            Digital Autonomy · Prototype
          </p>
          <h1 className="text-2xl font-semibold">Make a payment</h1>
        </div>
      </header>

      <Alert className="border-emerald-200 bg-emerald-50 text-emerald-950 [&>svg]:text-emerald-700">
        <ShieldCheck />
        <AlertTitle>Demo only</AlertTitle>
        <AlertDescription className="text-emerald-900">
          This uses a mock banking API. No real money moves.
        </AlertDescription>
      </Alert>

      <form className="grid gap-6 md:grid-cols-2" onSubmit={submitPayment}>
        <Card>
          <CardHeader>
            <div>
              <CardDescription>
                Payment details
              </CardDescription>
              <CardTitle>Who is this payment for?</CardTitle>
            </div>
            <Badge variant="outline">GBP £</Badge>
          </CardHeader>
          <CardContent className="space-y-4">
            <PersonPicker
              title="Paying from"
              name="fromUser"
              users={users}
              selectedUser={fromUser}
              onChange={changeSender}
            />

            <ArrowDown className="mx-auto size-4 text-muted-foreground" aria-hidden="true" />

            <PersonPicker
              title="Sending to"
              name="toUser"
              users={users}
              selectedUser={toUser}
              disabledUser={fromUser}
              onChange={changeRecipient}
            />

            <div className="space-y-2">
              <Label htmlFor="amount">Amount (GBP)</Label>
              <Input
                id="amount"
                name="amount"
                type="text"
                inputMode="decimal"
                autoComplete="off"
                placeholder="0.00"
                value={amount}
                onChange={(event) => {
                  setAmount(event.target.value)
                  setResult(null)
                  setError('')
                }}
                aria-describedby="amount-help"
                required
              />
              <p className="text-sm text-muted-foreground" id="amount-help">
                Enter an amount in pounds, up to two decimal places.
              </p>
            </div>

            {error && (
              <Alert variant="destructive">
                <CircleAlert />
                <AlertTitle>Payment not sent</AlertTitle>
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            <Button type="submit" disabled={submitting} className="w-full">
              {submitting ? (
                <><LoaderCircle className="size-4 animate-spin" /> Sending demo payment…</>
              ) : (
                <>Review and send <ArrowRight /></>
              )}
            </Button>
            <p className="flex items-center justify-center gap-2 text-sm text-muted-foreground">
              <ShieldCheck className="size-4" />
              Your demo details are sent to the local API only.
            </p>
          </CardContent>
        </Card>

        <aside className="space-y-4" aria-label="Payment summary and result">
          <Card>
            <CardHeader>
              <CardDescription>Summary</CardDescription>
              <CardTitle>Your payment</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {[{ user: sender, label: 'Paying from' }, { user: recipient, label: 'Sending to' }].map(({ user, label }, index) => (
                <div key={`${label}-${user.id}`}>
                  {index > 0 && (
                    <div className="mb-4 ml-5 h-5 border-l" aria-hidden="true" />
                  )}
                  <div className="flex items-center gap-3">
                    <Avatar>
                      <AvatarFallback className={user.id === 'alice' ? 'bg-emerald-100 text-emerald-800' : 'bg-sky-100 text-sky-800'}>
                        {user.initials}
                      </AvatarFallback>
                    </Avatar>
                    <div>
                      <p className="text-xs text-muted-foreground">{label}</p>
                      <p className="font-medium">{user.name}</p>
                      <p className="text-sm text-muted-foreground">{user.account}</p>
                    </div>
                  </div>
                </div>
              ))}
              <Separator />
              <div className="flex items-center justify-between">
                <span className="text-sm text-muted-foreground">Amount</span>
                <span className="font-semibold">£{formattedAmount}</span>
              </div>
            </CardContent>
          </Card>

          {result && (
            <Card aria-live="polite">
              <CardHeader>
                <div className="flex items-center gap-3">
                  <span className="grid size-9 place-items-center rounded-full bg-emerald-100 text-emerald-800">
                    <Check className="size-5" />
                  </span>
                  <div>
                    <CardDescription>Mock API response</CardDescription>
                    <CardTitle>Payment request sent</CardTitle>
                  </div>
                </div>
              </CardHeader>
              <CardContent className="space-y-4">
                <p className="text-xl font-semibold">£{formattedAmount} GBP</p>
                <div className="flex items-center gap-2 text-muted-foreground">
                  {sender.name} <ArrowRight className="size-4" /> {recipient.name}
                </div>
                <Separator />
                <dl className="space-y-2">
                  <div className="flex justify-between gap-3">
                    <dt className="text-muted-foreground">Payment status</dt>
                    <dd><Badge variant="secondary">{result.payment?.Data?.Status || 'Submitted'}</Badge></dd>
                  </div>
                  <div className="flex justify-between gap-3">
                    <dt className="text-muted-foreground">Consent status</dt>
                    <dd><Badge variant="secondary">{result.consent?.Data?.Status || 'Created'}</Badge></dd>
                  </div>
                  {result.payment?.Data?.DomesticPaymentId && (
                    <div className="flex justify-between gap-3">
                      <dt className="text-muted-foreground">Payment reference</dt>
                      <dd className="max-w-40 truncate text-right font-medium">{result.payment.Data.DomesticPaymentId}</dd>
                    </div>
                  )}
                </dl>
                <details>
                  <summary className="cursor-pointer font-medium text-primary">View full API response</summary>
                  <pre className="mt-3 max-h-60 overflow-auto rounded-md bg-muted p-3 text-xs">
                    {JSON.stringify(result, null, 2)}
                  </pre>
                </details>
                <p className="text-sm text-muted-foreground">
                  This is a mock response; the status does not confirm a real transfer.
                </p>
              </CardContent>
            </Card>
          )}
        </aside>
      </form>

      <footer className="text-center text-sm text-muted-foreground">
        Prototype environment · Payments are simulated and do not move real money
      </footer>
    </main>
  )
}

export default App
