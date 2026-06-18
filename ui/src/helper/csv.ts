import { exportFile, Notify, type QTableColumn } from 'quasar';
import type { ComposerTranslation } from 'vue-i18n';

interface CSVColumns {
  field?: string | ((row: Record<string, unknown>) => string);
  label: string;
  name?: string;
}

/**
 * Encode an array of objects into a CSV string.
 */
function csvEncoder(
  objects: Record<string, unknown>[],
  csvColumns: CSVColumns[],
): string {
  const separator = ';';
  const columns = csvColumns.map((col) => col.label);
  const csvRows = objects.map((obj) => {
    const rowElements = csvColumns.map((col) => {
      if (typeof col.field === 'function') {
        const val = col.field(obj);
        if (val !== undefined) {
          return val;
        }
      }
      const key = (col.field as string) ?? col.name;
      return Object.hasOwn(obj, key) ? obj[key] : '';
    });
    return rowElements.join(separator);
  });
  return [
    columns.join(separator),
    csvRows.join('\r\n'),
  ].join('\r\n');
}

/**
 * Export an array of objects as CSV.
 */
function exportAsCsv(objects: Record<string, unknown>[], csvColumns: CSVColumns[]): boolean|Error {
  const content = csvEncoder(objects, csvColumns);
  const status = exportFile('table-export.csv', content, 'text/csv');
  if (status !== true) {
    console.error('Error during file download', status);
  }
  return status;
}


/**
 * Util function to export a QTable as CSV.
 * Typesafe and notify on error.
 */
export function exportQTableAsCsv(rows: object[], columns: QTableColumn[], t: ComposerTranslation): boolean|Error {
  const objects = rows.map((node) => {
    return Object.entries(node).reduce(
      (acc, [key, value]) => {
        acc[key] = value;
        return acc;
      },
      {} as Record<string, unknown>,
    );
  });
  const status = exportAsCsv(objects, columns);
  if (status !== true) {
    Notify.create({
      type:'negative',
      message: t('ERROR_CSV_EXPORT'),
    });
  }
  return status;
}
