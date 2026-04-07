import { createI18n } from 'vue-i18n';
import en from './locales/en.json';
import nl from './locales/nl.json';
import fr from './locales/fr.json';

const i18n = createI18n({
  legacy: false, // you must set `false`, to use Composition API
  globalInjection: true,
  locale: 'EN', // set locale
  fallbackLocale: 'EN', // set fallback locale
  messages: {
    EN: en,
    NL: nl,
    FR: fr
  }
});

export default i18n;
