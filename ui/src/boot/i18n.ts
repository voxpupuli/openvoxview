import { defineBoot } from '#q-app/wrappers';
import { createI18n } from 'vue-i18n';

import messages from 'src/i18n';

export type MessageLanguages = keyof typeof messages;
// Type-define 'en-US' as the master schema for the resource
export type MessageSchema = (typeof messages)['en-US'];

// See https://vue-i18n.intlify.dev/guide/advanced/typescript.html#global-resource-schema-type-definition
/* eslint-disable @typescript-eslint/no-empty-object-type */
declare module 'vue-i18n' {
  // define the locale messages schema
  export interface DefineLocaleMessage extends MessageSchema {}

  // define the datetime format schema
  export interface DefineDateTimeFormat {}

  // define the number format schema
  export interface DefineNumberFormat {}
}
/* eslint-enable @typescript-eslint/no-empty-object-type */
export default defineBoot(({ app }) => {
  const i18n = createI18n<{ message: MessageSchema }, MessageLanguages>({
    locale: getBrowserLocale(),
    legacy: false,
    messages,
  });

  // Set i18n instance on app
  app.use(i18n);
});

export function getBrowserLocale(): MessageLanguages {
  const availableLocales: MessageLanguages[] = Object.keys(messages) as MessageLanguages[];
  const fallbackLocale: MessageLanguages = 'en-US';

  if (navigator.languages) {
    for (const lang of navigator.languages) {
      // Try an exact match
      if (availableLocales.includes(lang as MessageLanguages)) {
        return lang as MessageLanguages;
      }

      // Otherwise try to split language code and try to compare the first part only
      const langCode = lang.split('-')[0];
      const match = availableLocales.find(
        (loc) => loc.split('-')[0] === langCode,
      );
      if (match) {
        return match;
      }
    }
  }

  return fallbackLocale;
}
