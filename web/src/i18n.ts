export type Lang = 'en' | 'ru'

export interface Strings {
  nav_repos: string
  nav_settings: string
  stat_repos: string
  stat_tags: string
  stat_storage: string
  signout: string
  catalog_title: string
  filter_ph: string
  empty_title: string
  empty_sub: string
  clear_filter: string
  lbl_tags: string
  lbl_size: string
  lbl_updated: string
  copy: string
  copied: string
  th_tag: string
  th_digest: string
  th_size: string
  th_created: string
  th_actions: string
  view_image: string
  delete_tag: string
  copy_digest: string
  platforms_list: string
  total_size: string
  created: string
  platform: string
  layers: string
  configuration: string
  palette_ph: string
  no_matches: string
  type_repo: string
  type_tag: string
  confirm_title: string
  confirm_sub: string
  confirm_warn: string
  cancel: string
  login_signin: string
  login_connect: string
  login_user: string
  login_pass: string
  login_url: string
  login_note: string
  toast_copied: string
  toast_405: string
  toast_deleted: string
  settings_title: string
  settings_sub: string
  settings_sec_general: string
  settings_appearance: string
  settings_appearance_sub: string
  settings_language: string
  settings_language_sub: string
  settings_sec_about: string
  settings_github: string
  settings_github_sub: string
  settings_github_link: string
  theme_light: string
  theme_dark: string
  multi_arch: string
  loading: string
  load_error: string
}

export const T: Record<Lang, Strings> = {
  en: {
    nav_repos: 'Repositories',
    nav_settings: 'Settings',
    stat_repos: 'Repositories',
    stat_tags: 'Tags',
    stat_storage: 'Storage',
    signout: 'Sign out',
    catalog_title: 'Repositories',
    filter_ph: 'Filter by name…',
    empty_title: 'No repositories match',
    empty_sub: 'Try a different name, or clear the filter.',
    clear_filter: 'Clear filter',
    lbl_tags: 'tags',
    lbl_size: 'size',
    lbl_updated: 'updated',
    copy: 'Copy',
    copied: 'Copied',
    th_tag: 'Tag',
    th_digest: 'Digest',
    th_size: 'Size',
    th_created: 'Created',
    th_actions: 'Actions',
    view_image: 'View image',
    delete_tag: 'Delete tag',
    copy_digest: 'Copy digest',
    platforms_list: 'Platforms · manifest list',
    total_size: 'Total size',
    created: 'Created',
    platform: 'Platform',
    layers: 'Layers',
    configuration: 'Configuration',
    palette_ph: 'Search repositories and tags…',
    no_matches: 'No matches',
    type_repo: 'Repo',
    type_tag: 'Tag',
    confirm_title: 'Delete this tag?',
    confirm_sub: 'Sends DELETE manifest by digest.',
    confirm_warn:
      'This unlinks the manifest only. Disk space is reclaimed when registry garbage-collect runs (read-only mode, CLI/cronjob — not a button).',
    cancel: 'Cancel',
    login_signin: 'Sign in',
    login_connect: 'Connect to',
    login_user: 'Username',
    login_pass: 'Password',
    login_url: 'Registry',
    login_note:
      'Auth follows the token flow: a 401 returns a WWW-Authenticate realm; we exchange credentials for a bearer token scoped per repository & action.',
    toast_copied: 'Copied to clipboard',
    toast_405:
      '405 UNSUPPORTED — deletion is disabled on this registry (set REGISTRY_STORAGE_DELETE_ENABLED=true).',
    toast_deleted: 'Manifest deleted — disk space is reclaimed only after garbage-collect.',
    settings_title: 'Settings',
    settings_sub: 'Appearance, language and registry behavior.',
    settings_sec_general: 'General',
    settings_appearance: 'Theme',
    settings_appearance_sub: 'Light or dark interface.',
    settings_language: 'Language',
    settings_language_sub: 'Interface language.',
    settings_sec_about: 'About',
    settings_github: 'Source code',
    settings_github_sub: 'View the project on GitHub.',
    settings_github_link: 'Open GitHub',
    theme_light: 'Light',
    theme_dark: 'Dark',
    multi_arch: 'multi-arch',
    loading: 'Loading…',
    load_error: 'Failed to load',
  },
  ru: {
    nav_repos: 'Репозитории',
    nav_settings: 'Настройки',
    stat_repos: 'Репозитории',
    stat_tags: 'Теги',
    stat_storage: 'Хранилище',
    signout: 'Выйти',
    catalog_title: 'Репозитории',
    filter_ph: 'Фильтр по имени…',
    empty_title: 'Ничего не найдено',
    empty_sub: 'Измените запрос или сбросьте фильтр.',
    clear_filter: 'Сбросить фильтр',
    lbl_tags: 'тегов',
    lbl_size: 'размер',
    lbl_updated: 'обновлён',
    copy: 'Копировать',
    copied: 'Скопировано',
    th_tag: 'Тег',
    th_digest: 'Дайджест',
    th_size: 'Размер',
    th_created: 'Создан',
    th_actions: 'Действия',
    view_image: 'Просмотр образа',
    delete_tag: 'Удалить тег',
    copy_digest: 'Копировать дайджест',
    platforms_list: 'Платформы · список манифестов',
    total_size: 'Общий размер',
    created: 'Создан',
    platform: 'Платформа',
    layers: 'Слои',
    configuration: 'Конфигурация',
    palette_ph: 'Поиск репозиториев и тегов…',
    no_matches: 'Ничего не найдено',
    type_repo: 'Репо',
    type_tag: 'Тег',
    confirm_title: 'Удалить этот тег?',
    confirm_sub: 'Отправит DELETE manifest по дайджесту.',
    confirm_warn:
      'Открепляется только манифест. Место освобождается при запуске registry garbage-collect (read-only режим, CLI/cronjob — не кнопка).',
    cancel: 'Отмена',
    login_signin: 'Войти',
    login_connect: 'Подключение к',
    login_user: 'Имя пользователя',
    login_pass: 'Пароль',
    login_url: 'Registry',
    login_note:
      'Авторизация по token-флоу: 401 возвращает realm в WWW-Authenticate; обмениваем учётные данные на bearer-токен со скоупом на репозиторий и действие.',
    toast_copied: 'Скопировано в буфер обмена',
    toast_405:
      '405 UNSUPPORTED — удаление отключено на этом registry (нужно REGISTRY_STORAGE_DELETE_ENABLED=true).',
    toast_deleted: 'Манифест удалён — место освободится только после garbage-collect.',
    settings_title: 'Настройки',
    settings_sub: 'Внешний вид, язык и поведение registry.',
    settings_sec_general: 'Основные',
    settings_appearance: 'Тема',
    settings_appearance_sub: 'Светлый или тёмный интерфейс.',
    settings_language: 'Язык',
    settings_language_sub: 'Язык интерфейса.',
    settings_sec_about: 'О проекте',
    settings_github: 'Исходный код',
    settings_github_sub: 'Открыть проект на GitHub.',
    settings_github_link: 'Открыть GitHub',
    theme_light: 'Светлая',
    theme_dark: 'Тёмная',
    multi_arch: 'multi-arch',
    loading: 'Загрузка…',
    load_error: 'Не удалось загрузить',
  },
}
