import { autoImplement } from 'src/helper/functions';

type PuppetEventCountSubject = {
  title: string;
  type: string | null;
};

export interface ApiPuppetEventCount {
  failures: number;
  skips: number;
  successes: number;
  noops: number;
  subject_type: string;
  subject: PuppetEventCountSubject;
}

export class PuppetEventCount extends autoImplement<ApiPuppetEventCount>() {
  static fromApi(apiItem: ApiPuppetEventCount): PuppetEventCount {
    return new PuppetEventCount(apiItem);
  }
}
