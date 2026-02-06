import moment from 'moment/moment';

export function autoImplement<T extends object>(defaults: Partial<T> = {}) {
  return class {
    constructor(data: Partial<T> = {}) {
      Object.assign(this, defaults);
      Object.assign(this, data);
    }
  } as new (data?: T) => T;
}

export function formatTimestamp(input: string|Date|null, withMs?: boolean) : string | null {
  if (!input) return null;
  return moment(input).format(`DD. MMM. YYYY - HH:mm:ss${withMs ? '.SSS' : ''}`)
}

export function formatDuration(milliseconds: number): string {
  if (milliseconds < 0) return "0 ms";

  const units = [
    { name: "d", value: 24 * 60 * 60 * 1000 },
    { name: "h", value: 60 * 60 * 1000 },
    { name: "m", value: 60 * 1000 },
    { name: "s", value: 1000 },
    { name: "ms", value: 1 }
  ];

  const parts: string[] = [];
  let remaining = milliseconds;

  for (const unit of units) {
    if (remaining >= unit.value) {
      const count = Math.floor(remaining / unit.value);
      parts.push(`${count} ${unit.name}`);
      remaining %= unit.value;
    }
  }

  return parts.length > 0 ? parts.join(" ") : "0 ms";
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

export async function copyToClipboard(payload: any): Promise<void> {
  await navigator.clipboard.writeText(payload);
}
