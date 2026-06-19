import { useApp } from '@/state/app'
import { SunIcon, MoonIcon } from '@/components/icons'

export function Settings() {
  const { t, theme, setTheme, lang, setLang } = useApp()

  return (
    <div className="fade settings-narrow">
      <h1 className="page-title">{t.settings_title}</h1>
      <p className="page-sub">{t.settings_sub}</p>

      <div className="section-label">{t.settings_sec_general}</div>
      <div className="card flush">
        <div className="setting-row">
          <div>
            <div className="setting-name">{t.settings_appearance}</div>
            <div className="setting-desc">{t.settings_appearance_sub}</div>
          </div>
          <div className="seg">
            <button
              className={`seg-btn${theme === 'light' ? ' active' : ''}`}
              onClick={() => setTheme('light')}
            >
              <SunIcon size={13} />
              {t.theme_light}
            </button>
            <button
              className={`seg-btn${theme === 'dark' ? ' active' : ''}`}
              onClick={() => setTheme('dark')}
            >
              <MoonIcon size={13} />
              {t.theme_dark}
            </button>
          </div>
        </div>

        <div className="setting-row">
          <div>
            <div className="setting-name">{t.settings_language}</div>
            <div className="setting-desc">{t.settings_language_sub}</div>
          </div>
          <div className="seg">
            <button
              className={`seg-btn lang${lang === 'en' ? ' active' : ''}`}
              onClick={() => setLang('en')}
            >
              English
            </button>
            <button
              className={`seg-btn lang${lang === 'ru' ? ' active' : ''}`}
              onClick={() => setLang('ru')}
            >
              Русский
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
