export interface BaseResponse<T> {
  Data: T;
}

export interface ErrorResponse {
  Error: string
}

export interface ApiMeta {
  CaEnabled: boolean
  CaReadOnly: boolean
  UnreportedHours: number
  StripPathPrefix: string
  AuthEnabled: boolean
}

export interface ApiVersion {
  Version: string;
}

export interface LoginResponse {
  access_token: string
  refresh_token: string
  expires_in: number
}

export interface UserProfile {
  id: number
  username: string
  email: string
  display_name: string
  given_name: string
  surname: string
  auth_source: string
  created_at: string
  updated_at: string
}

export interface CreateUserRequest {
  username: string
  password: string
  email?: string
  display_name?: string
}

export interface UpdateUserRequest {
  email?: string
  display_name?: string
  password?: string
}
