import { autoImplement } from 'src/helper/functions';

export enum CertificateState {
  CertificateRequested = "requested",
  CertificateSigned = "signed",
  CertificateRevoked = "revoked",
}

export type CertificateStatus = {
  name: string;
  state: CertificateState;
  fingerprint: string;
  fingerprints: Record<string, string>;
  dns_alt_names: string[];
  subject_alt_names?: string[];
  serial_number?: string;
  authorization_extensions?: Record<string, string>;
  not_before?: Date;
  not_after?: Date;
}

export type CertificateStatusQuery = {
  states?: CertificateState[];
  filter?: string;
}

type ApiCertificateStatusResponse = {
  certificate_statuses?: CertificateStatus[];
}

export class CertificateStatusResponse extends autoImplement<ApiCertificateStatusResponse>() {
  static fromApi(apiItem: ApiCertificateStatusResponse): CertificateStatusResponse {
    return new CertificateStatusResponse(apiItem);
  }
}
