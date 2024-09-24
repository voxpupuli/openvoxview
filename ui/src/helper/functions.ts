import moment from 'moment/moment';

export function autoImplement<T extends object>(defaults: Partial<T> = {}) {
  // eslint-disable-next-line @typescript-eslint/no-extraneous-class
  return class {
    constructor(data: Partial<T> = {}) {
      Object.assign(this, defaults);
      Object.assign(this, data);
    }
  } as new (data?: T) => T;
}

export function formatTimestamp(input: string|Date, withMs?: boolean) : string {
  return moment(input).format(`DD. MMM. YYYY - HH:mm:ss${withMs ? '.SSS' : ''}`)
}

export function msToTime(duration: number) {
  const milliseconds = (duration%1000)/100
    , seconds = parseInt(((duration/1000)%60).toString())
    , minutes = parseInt(((duration/(1000*60))%60).toString())
    , hours = parseInt(((duration/(1000*60*60))%24).toString());

  const hoursStr = (hours < 10) ? '0' + hours : hours.toString();
  const minutesStr = (minutes < 10) ? '0' + minutes : minutes.toString();
  const secondsStr  = (seconds < 10) ? '0' + seconds : seconds.toString();

  return `${hoursStr}:${minutesStr}:${secondsStr}.${milliseconds.toString().replace('.','')}`;
}

/* eslint-disable  @typescript-eslint/no-explicit-any */
export function getOsNameFromOsFact(os_fact: any) : string {
  switch (os_fact['family']) {
    case 'windows':
      return os_fact['windows']['product_name']
    case 'Darwin':
      return os_fact['macosx']['product']
    default:
      return os_fact['distro']['description']
  }
}
