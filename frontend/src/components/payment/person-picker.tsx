import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'

export type DemoUser = {
  id: string
  name: string
  initials: string
  account: string
}

type PersonPickerProps = {
  title: string
  name: string
  users: DemoUser[]
  selectedUser: string
  disabledUser?: string
  onChange: (userId: string) => void
}

export function PersonPicker({
  title,
  name,
  users,
  selectedUser,
  disabledUser,
  onChange,
}: PersonPickerProps) {
  return (
    <fieldset className="space-y-2">
      <legend className="mb-2 text-sm font-medium">{title}</legend>
            <RadioGroup value={selectedUser} onValueChange={onChange} className="gap-2">
        {users.map((user) => {
          const optionId = `${name}-${user.id}`
          const isSelected = selectedUser === user.id
          const isDisabled = disabledUser === user.id

          return (
            <div
              className={`flex items-center gap-3 rounded-lg border p-3 transition-colors ${
                isSelected ? 'border-emerald-600 bg-emerald-50' : ''
              } ${isDisabled ? 'opacity-50' : ''}`}
              key={optionId}
            >
              <RadioGroupItem
                id={optionId}
                value={user.id}
                disabled={isDisabled}
                aria-label={`${user.name}, ${user.account}`}
              />
              <Avatar>
                <AvatarFallback className={user.id === 'alice' ? 'bg-emerald-100 text-emerald-800' : 'bg-sky-100 text-sky-800'}>
                  {user.initials}
                </AvatarFallback>
              </Avatar>
              <Label htmlFor={optionId} className="min-w-0 flex-1 cursor-pointer">
                <span className="block font-medium">{user.name}</span>
                <span className="block text-sm text-muted-foreground">
                  {user.account}
                </span>
              </Label>
            </div>
          )
        })}
      </RadioGroup>
    </fieldset>
  )
}
