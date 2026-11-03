export type DiffType = 'same' | 'add' | 'del';
export type DiffStyle = 'empty' | 'normal' | 'removed' | 'added';

export type DiffRow<T> = {
  type: DiffType;
  oldLine: number | null;
  newLine: number | null;
  oldItem: T | null;
  newItem: T | null;
};

export type DiffItem<T> = {
  style: DiffStyle;
  item: T | null;
};


export function diffObjectsWithoutContext<T>(
  a: T[],
  b: T[],
  keyOf: (item: T) => string
): DiffRow<T>[] {
  const bPositions = new Map<string, number[]>();
  b.forEach((item, idx) => {
    const k = keyOf(item);
    const arr = bPositions.get(k) ?? [];
    arr.push(idx + 1);
    bPositions.set(k, arr);
  });

  const rows: DiffRow<T>[] = [];
  const usedB = new Set<number>();

  // walk a => same or del
  a.forEach((item, idx) => {
    const k = keyOf(item);
    const queue = bPositions.get(k);

    if (queue && queue.length > 0) {
      const newLine = queue.shift()!;
      usedB.add(newLine);
      rows.push({
        type: 'same',
        oldLine: idx + 1,
        newLine,
        oldItem: item,
        newItem: b[newLine - 1] ?? null,
      });
    } else {
      rows.push({
        type: 'del',
        oldLine: idx + 1,
        newLine: null,
        oldItem: item,
        newItem: null,
      });
    }
  });

  // remaining b => add
  b.forEach((item, idx) => {
    const newLine = idx + 1;
    if (!usedB.has(newLine)) {
      rows.push({
        type: 'add',
        oldLine: null,
        newLine,
        oldItem: null,
        newItem: item,
      });
    }
  });

  return rows;
}
