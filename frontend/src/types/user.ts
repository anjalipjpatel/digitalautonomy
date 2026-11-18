export const USER_ROLES = {
    ADMIN: 'admin',
    OWNER: 'vulnerable_owner',
    DELEGATE: 'delegate'
} as const

export type UserRole = (typeof USER_ROLES)[keyof typeof USER_ROLES]