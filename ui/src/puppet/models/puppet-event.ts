import { autoImplement } from 'src/helper/functions';

export interface ApiPuppetEvent {
  certname: string;
  configuration_version: string;
  containing_class: string;
  containment_path: string[];
  corrective_change: null;
  environment: string;
  file: string;
  line: number;
  message: string;
  name: string;
  new_value: unknown;
  old_value: unknown;
  property: null;
  report: string;
  report_receive_time: Date;
  resource_title: string;
  resource_type: string;
  run_end_time: Date;
  run_start_time: Date;
  status: string;
  timestamp: Date;
}

export class PuppetEvent extends autoImplement<ApiPuppetEvent>() {
  static fromApi(apiItem: ApiPuppetEvent) : PuppetEvent {
    return new PuppetEvent(apiItem);
  }
}
