export const USER_ROLES = {
    ADMIN: 'admin',
    OWNER: 'vulnerable_owner',
    DELEGATE: 'delegate'
} as const

export type UserRole = (typeof USER_ROLES)[keyof typeof USER_ROLES]

export interface UserProfile {
    ID: string
    Email: string
    Name: string
    Role: UserRole
    CreatedAt: string
}