import { autoImplement } from 'src/helper/functions';

export interface ApiPuppetCatalog {
  certname: string;
  version: string;
  environment: string;
  hash: string;
  transaction_uuid: string;
  code_id: string;
  producer_timestamp: string; // time.Time -> string (ISO timestamp)
  producer: string;
  resources: CatalogResource;
  edges: CatalogEdge;
}

export interface CatalogResource {
  href: string;
  data: CatalogResourceData[];
}

export interface CatalogResourceData {
  certname: string;
  resource: string;
  type: string;
  title: string;
  exported: boolean;
  tags: string[];
  file: string;
  line: number;
  parameters: unknown;
}

export interface CatalogEdge {
  href: string;
  data: CatalogEdgeData[];
}

export interface CatalogEdgeData {
  relationship: string;
  source_title: string;
  source_type: string;
  target_title: string;
  target_type: string;
}

export class PuppetCatalog extends autoImplement<ApiPuppetCatalog>() {
  static fromApi(apiItem: ApiPuppetCatalog): PuppetCatalog {
    return new PuppetCatalog(apiItem);
  }
}
