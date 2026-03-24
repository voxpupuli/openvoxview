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
}

export interface ApiVersion {
  Version: string;
}
