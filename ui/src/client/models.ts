export interface BaseResponse<T> {
  Data: T;
}

export interface ErrorResponse {
  Error: string
}

export interface ApiMeta {
  UnreportedHours: number
}

export interface ApiVersion {
  Version: string;
}
