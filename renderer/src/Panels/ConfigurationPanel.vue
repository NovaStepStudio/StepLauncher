<template>
  <div class="config-panel">
    <div class="config-sidebar">
      <button class="back-btn" @click="$emit('back')">
        <img :src="'assets/svg/arrow-left.svg'" width="16" height="16" />
      </button>
      <div class="sidebar-sections">
        <button
          v-for="s in sections"
          :key="s.key"
          :class="['section-btn', { active: section === s.key }]"
          @click="section = s.key"
        >
          <img class="section-icon" :src="s.icon" width="15" height="15" />
          <span class="section-label">{{ s.label }}</span>
        </button>
      </div>
    </div>

    <div class="config-content">
      <transition name="slide" mode="out-in">
        <div v-if="section === 'launcher'" key="launcher" class="section-content">
          <h2 class="content-title">{{ $t('configuration.launcher.title') }}</h2>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.launcher.auto_start_minecraft') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.launcher.autoStartMinecraft" @change="saveLauncher('autoStartMinecraft')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.launcher.hide_on_launch') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.launcher.hideOnLaunch" @change="saveLauncher('hideOnLaunch')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.launcher.show_console') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.launcher.showConsole" @change="saveLauncher('showConsole')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.launcher.show_news') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.launcher.showNews" @change="saveLauncher('showNews')" />
              <span class="slider" />
            </label>
          </div>

          <button class="btn-launcher-folder" @click="openLauncherFolder"><img :src="'assets/svg/folder.svg'" width="14" height="14" /> {{ $t('configuration.launcher.open_launcher_folder') }}</button>

          <h3 class="subsection-title">{{ $t('configuration.launcher.subsection_performance') }}</h3>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.launcher.filters') }}</span>
              <span class="setting-desc">{{ $t('configuration.launcher.filters_desc') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.launcher.filters" @change="saveLauncher('filters')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.launcher.blur') }}</span>
              <span class="setting-desc">{{ $t('configuration.launcher.blur_desc') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.launcher.blur" @change="saveLauncher('blur')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.launcher.hardware_acceleration') }}</span>
              <span class="setting-desc">{{ $t('configuration.launcher.hardware_acceleration_desc') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.launcher.hardwareAcceleration" @change="saveLauncher('hardwareAcceleration')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.launcher.shadows') }}</span>
              <span class="setting-desc">{{ $t('configuration.launcher.shadows_desc') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.launcher.shadows" @change="saveLauncher('shadows')" />
              <span class="slider" />
            </label>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.launcher.subsection_integration') }}</h3>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.launcher.discord_rpc') }}</span>
              <span class="setting-desc">{{ $t('configuration.launcher.discord_rpc_desc') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.launcher.discordRpc" @change="saveLauncher('discordRpc')" />
              <span class="slider" />
            </label>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.language.title') }}</h3>

          <div class="lang-grid">
            <button
              v-for="loc in locales"
              :key="loc"
              :class="['lang-btn', { active: locale === loc }]"
              @click="changeLanguage(loc)"
            >
              <img :src="getFlagIcon(loc)" width="28" height="20" class="lang-flag" />
              <div class="lang-info">
                <div class="lang-name-row">
                  <span class="lang-name">{{ $t('configuration.language.names.lang_' + loc.replace('-', '_')) }}</span>
                  <span class="lang-badge">{{ $t('configuration.language.official') }}</span>
                </div>
                <div class="lang-level-row">
                  <span class="lang-level-bar">
                    <span class="lang-level-fill" :style="{ width: getLangMeta(loc).level + '%' }"></span>
                  </span>
                  <span class="lang-level-text">{{ getLangMeta(loc).level }}%</span>
                </div>
              </div>
            </button>
          </div>

        </div>

        <div v-else-if="section === 'minecraft'" key="minecraft" class="section-content">
          <h2 class="content-title">{{ $t('configuration.minecraft.title') }}</h2>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.use_recommended_java') }}</span>
              <span class="setting-desc">{{ $t('configuration.minecraft.use_recommended_java_desc') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.minecraft.useRecommendedJava" @change="saveMinecraft('useRecommendedJava')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.max_console_events') }}</span>
            </div>
            <input type="number" class="input-number" v-model.number="config.minecraft.maxConsoleEvents" @change="saveMinecraft('maxConsoleEvents')" min="10" max="10000" />
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.notify_on_launch') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.minecraft.showNotificationOnLaunch" @change="saveMinecraft('showNotificationOnLaunch')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.clean_before_launch') }}</span>
              <span class="setting-desc">{{ $t('configuration.minecraft.clean_before_launch_desc') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.minecraft.cleanBeforeLaunch" @change="saveMinecraft('cleanBeforeLaunch')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.java_path') }}</span>
              <span class="setting-desc">{{ $t('configuration.minecraft.java_path_desc') }}</span>
            </div>
            <div class="input-with-btn">
              <input type="text" class="input-text" v-model="config.minecraft.javaPath" @change="saveMinecraft('javaPath')" :placeholder="$t('configuration.minecraft.java_path_placeholder')" />
              <button class="btn-browse" @click="browseJava">
                <img :src="'assets/svg/folder.svg'" width="14" height="14" />
              </button>
            </div>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.fullscreen') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.minecraft.fullscreen" @change="saveMinecraft('fullscreen')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.window_size') }}</span>
            </div>
            <div class="input-group-row">
              <div class="input-group">
                <span class="input-label">{{ $t('configuration.minecraft.window_width') }}</span>
                <input type="number" class="input-number" v-model.number="config.minecraft.windowWidth" @change="saveMinecraft('windowWidth')" min="400" max="7680" />
              </div>
              <div class="input-group">
                <span class="input-label">{{ $t('configuration.minecraft.window_height') }}</span>
                <input type="number" class="input-number" v-model.number="config.minecraft.windowHeight" @change="saveMinecraft('windowHeight')" min="300" max="4320" />
              </div>
            </div>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.ram_limit') }}</span>
              <span class="setting-desc">{{ $t('configuration.minecraft.ram_limit_desc') }}</span>
            </div>
            <div class="ram-control">
              <input type="range" class="ram-slider" v-model.number="config.minecraft.maxRam" @input="saveMinecraft('maxRam')" :min="512" :max="maxRam" step="256" />
              <span class="ram-value">{{ formatRam(config.minecraft.maxRam) }}</span>
            </div>
          </div>
          <div class="ram-hint">
            {{ $t('configuration.minecraft.ram_hint', { total: formatRam(totalRam), recommended: formatRam(recommendedRam) }) }}
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.ram_min') }}</span>
              <span class="setting-desc">{{ $t('configuration.minecraft.ram_min_desc') }}</span>
            </div>
            <div class="ram-control">
              <input type="range" class="ram-slider" v-model.number="config.minecraft.minRam" @input="saveMinecraft('minRam')" :min="512" :max="config.minecraft.maxRam" step="256" />
              <span class="ram-value">{{ formatRam(config.minecraft.minRam) }}</span>
            </div>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.gc_preset') }}</span>
              <span class="setting-desc">{{ $t('configuration.minecraft.gc_preset_desc') }}</span>
            </div>
            <div class="select-wrap">
              <select class="input-select" v-model="config.minecraft.gcPreset" @change="saveMinecraft('gcPreset')">
                <option value="">{{ $t('configuration.minecraft.gc_options.default') }}</option>
                <option value="serial">{{ $t('configuration.minecraft.gc_options.serial') }}</option>
                <option value="parallel">{{ $t('configuration.minecraft.gc_options.parallel') }}</option>
                <option value="g1gc">{{ $t('configuration.minecraft.gc_options.g1gc') }}</option>
                <option value="g1gc_optimized">{{ $t('configuration.minecraft.gc_options.g1gc_optimized') }}</option>
                <option value="zgc">{{ $t('configuration.minecraft.gc_options.zgc') }}</option>
                <option value="shenandoah">{{ $t('configuration.minecraft.gc_options.shenandoah') }}</option>
                <option value="epsilon">{{ $t('configuration.minecraft.gc_options.epsilon') }}</option>
              </select>
            </div>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.minecraft.gpu_preference') }}</span>
              <span class="setting-desc">{{ $t('configuration.minecraft.gpu_preference_desc') }}</span>
            </div>
            <div class="select-wrap">
              <select class="input-select" v-model="config.minecraft.gpuPreference" @change="saveMinecraft('gpuPreference')">
                <option value="">{{ $t('configuration.minecraft.gpu_options.none') }}</option>
                <option value="auto">{{ $t('configuration.minecraft.gpu_options.auto') }}</option>
                <option value="integrated">{{ $t('configuration.minecraft.gpu_options.integrated') }}</option>
                <option value="dedicated">{{ $t('configuration.minecraft.gpu_options.dedicated') }}</option>
              </select>
            </div>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.minecraft.subsection_advanced') }}</h3>

          <div class="args-card">
            <div class="args-card-header">
              <div class="args-card-info">
                <span class="args-card-label">{{ $t('configuration.minecraft.jvm_args_label') }}</span>
                <span class="args-card-desc">{{ $t('configuration.minecraft.jvm_args_desc') }}</span>
              </div>
              <span class="args-badge">{{ $t('configuration.minecraft.jvm_badge') }}</span>
            </div>
            <textarea class="args-editor" v-model="config.minecraft.jvmArgs" @change="saveMinecraft('jvmArgs')" :placeholder="$t('configuration.minecraft.jvm_args_placeholder')" rows="4" spellcheck="false"></textarea>
          </div>

          <div class="args-card">
            <div class="args-card-header">
              <div class="args-card-info">
                <span class="args-card-label">{{ $t('configuration.minecraft.game_args_label') }}</span>
                <span class="args-card-desc">{{ $t('configuration.minecraft.game_args_desc') }}</span>
              </div>
              <span class="args-badge args-badge--game">{{ $t('configuration.minecraft.game_badge') }}</span>
            </div>
            <textarea class="args-editor" v-model="config.minecraft.gameArgs" @change="saveMinecraft('gameArgs')" :placeholder="$t('configuration.minecraft.game_args_placeholder')" rows="3" spellcheck="false"></textarea>
          </div>
        </div>

        <div v-else-if="section === 'personalization'" key="personalization" class="section-content">
          <h2 class="content-title">{{ $t('configuration.personalization.title') }}</h2>

          <h3 class="subsection-title">{{ $t('configuration.personalization.subsection_colors') }}</h3>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.accent_color') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.accent_color_desc') }}</span>
            </div>
            <ColorPicker :model-value="config.personalization.accentColor" @update:model-value="config.personalization.accentColor = $event; savePersonalization('accentColor')" />
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.title_bar_color') }}</span>
            </div>
            <ColorPicker :model-value="config.personalization.titleBarColor" @update:model-value="config.personalization.titleBarColor = $event; savePersonalization('titleBarColor')" />
          </div>

          <h3 class="subsection-title">{{ $t('configuration.personalization.subsection_modals') }}</h3>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.modal_accent') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.modal_accent_desc') }}</span>
            </div>
            <ColorPicker :model-value="config.personalization.modalAccent ?? '#5cd0e7'" @update:model-value="config.personalization.modalAccent = $event; savePersonalization('modalAccent')" />
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.modal_background') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.modal_background_desc') }}</span>
            </div>
            <ColorPicker :model-value="config.personalization.modalBackground ?? '#1a1a30'" @update:model-value="config.personalization.modalBackground = $event; savePersonalization('modalBackground')" />
          </div>

          <div class="modal-preview-section">
            <span class="preview-label">{{ $t('configuration.personalization.preview_label') }}</span>
            <div class="modal-preview" :style="{ background: config.personalization.modalBackground ?? 'rgba(18,18,30,0.97)' }">
              <div class="mp-header" :style="{ color: config.personalization.modalAccent ?? '#5cd0e7' }">
                <img :src="'assets/svg/info.svg'" width="14" height="14" :style="{ stroke: config.personalization.modalAccent ?? '#5cd0e7' }" />
                <span>{{ $t('configuration.personalization.preview_dialog_title') }}</span>
                <span class="mp-close">&times;</span>
              </div>
              <div class="mp-body">
                <span class="mp-text" v-html="previewText"></span>
                <button class="mp-btn" :style="{ background: config.personalization.modalAccent ?? '#5cd0e7', color: '#000' }">{{ $t('configuration.personalization.preview_button') }}</button>
              </div>
            </div>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.personalization.subsection_panels') }}</h3>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.panel_background') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.panel_background_desc') }}</span>
            </div>
            <ColorPicker :model-value="config.personalization.panelBackground" @update:model-value="config.personalization.panelBackground = $event; savePersonalization('panelBackground')" />
          </div>

          <h3 class="subsection-title">{{ $t('configuration.personalization.subsection_sidebar') }}</h3>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.sidebar_background') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.sidebar_background_desc') }}</span>
            </div>
            <ColorPicker :model-value="config.personalization.sidebarBackground" @update:model-value="config.personalization.sidebarBackground = $event; savePersonalization('sidebarBackground')" />
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.sidebar_button_color') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.sidebar_button_color_desc') }}</span>
            </div>
            <ColorPicker :model-value="config.personalization.sidebarButtonColor" @update:model-value="config.personalization.sidebarButtonColor = $event; savePersonalization('sidebarButtonColor')" />
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.sidebar_position') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.sidebar_position_desc') }}</span>
            </div>
            <div class="pos-selector">
              <button :class="['pos-btn', { active: config.personalization.sidebarPosition === 'left' }]" @click="config.personalization.sidebarPosition = 'left'; savePersonalization('sidebarPosition')">
                <img :src="'assets/svg/layout-left.svg'" width="16" height="16" />
                {{ $t('configuration.personalization.sidebar_position_left') }}
              </button>
              <button :class="['pos-btn', { active: config.personalization.sidebarPosition === 'right' }]" @click="config.personalization.sidebarPosition = 'right'; savePersonalization('sidebarPosition')">
                <img :src="'assets/svg/layout-right.svg'" width="16" height="16" />
                {{ $t('configuration.personalization.sidebar_position_right') }}
              </button>
            </div>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.personalization.subsection_notifications') }}</h3>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.notif_error') }}</span>
            </div>
            <ColorPicker :model-value="config.personalization.notificationErrorColor" @update:model-value="config.personalization.notificationErrorColor = $event; savePersonalization('notificationErrorColor')" />
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.notif_warning') }}</span>
            </div>
            <ColorPicker :model-value="config.personalization.notificationWarnColor" @update:model-value="config.personalization.notificationWarnColor = $event; savePersonalization('notificationWarnColor')" />
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.notif_success') }}</span>
            </div>
            <ColorPicker :model-value="config.personalization.notificationSuccessColor" @update:model-value="config.personalization.notificationSuccessColor = $event; savePersonalization('notificationSuccessColor')" />
          </div>

          <h3 class="subsection-title">{{ $t('configuration.personalization.subsection_app_bg') }}</h3>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.app_background') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.app_background_desc') }}</span>
            </div>
            <div class="select-wrap" style="display:flex;gap:0.4rem;align-items:center">
              <select class="input-select" v-model="bgPreset" @change="onBgPreset" style="flex:1">
                <option v-for="n in 61" :key="n" :value="`url('assets/background/RRE36/${n}.webp')`">{{ $t('configuration.personalization.bg_option_rre36', { n }) }}</option>
              </select>
              <button class="btn-browse" @click="browseBackground" :title="$t('configuration.personalization.btn_upload')">
                <img :src="'assets/svg/upload.svg'" width="14" height="14" />
                {{ $t('configuration.personalization.btn_upload') }}
              </button>
            </div>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.personalization.subsection_typography') }}</h3>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.font_primary') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.font_primary_desc') }}</span>
            </div>
            <select class="input-select" v-model="config.personalization.fontPrimary" @change="savePersonalization('fontPrimary')">
              <option value="Lexend">{{ $t('configuration.personalization.font_lexend') }}</option>
              <option value="Inter">{{ $t('configuration.personalization.font_inter') }}</option>
            </select>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.font_secondary') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.font_secondary_desc') }}</span>
            </div>
            <select class="input-select" v-model="config.personalization.fontSecondary" @change="savePersonalization('fontSecondary')">
              <option value="Inter">{{ $t('configuration.personalization.font_inter') }}</option>
              <option value="Lexend">{{ $t('configuration.personalization.font_lexend') }}</option>
            </select>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.personalization.subsection_titlebar') }}</h3>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.macos_style') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.macos_style_desc') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.personalization.macOSTitlebar" @change="savePersonalization('macOSTitlebar')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.show_icon') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.show_icon_desc') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.personalization.showIcon" @change="savePersonalization('showIcon')" />
              <span class="slider" />
            </label>
          </div>

          <div class="setting">
            <div class="setting-info">
              <span class="setting-label">{{ $t('configuration.personalization.invert_position') }}</span>
              <span class="setting-desc">{{ $t('configuration.personalization.invert_position_desc') }}</span>
            </div>
            <label class="switch">
              <input type="checkbox" v-model="config.personalization.invertPosition" @change="savePersonalization('invertPosition')" />
              <span class="slider" />
            </label>
          </div>
        </div>

        <div v-else-if="section === 'themes'" key="themes" class="section-content">
          <h2 class="content-title">{{ $t('configuration.themes.title') }}</h2>

          <div class="theme-actions">
            <button class="theme-btn theme-btn--export" @click="showExportForm = true">
              <img :src="'assets/svg/upload.svg'" width="14" height="14" />
              {{ $t('configuration.themes.buttons.export_theme') }}
            </button>
            <button class="theme-btn theme-btn--import" @click="importTheme">
              <img :src="'assets/svg/download.svg'" width="14" height="14" />
              {{ $t('configuration.themes.buttons.import') }}
            </button>
          </div>

          <div v-if="showExportForm" class="export-form">
            <div class="export-field">
              <span class="export-label">{{ $t('configuration.themes.export.label_name') }}</span>
              <input v-model="exportName" class="export-input" :placeholder="$t('configuration.themes.export.placeholder_name')" />
            </div>
            <div class="export-field">
              <span class="export-label">{{ $t('configuration.themes.export.label_author') }}</span>
              <input v-model="exportAuthor" class="export-input" :placeholder="$t('configuration.themes.export.placeholder_author')" />
            </div>
            <div class="export-field">
              <span class="export-label">{{ $t('configuration.themes.export.label_version') }}</span>
              <input v-model="exportVersion" class="export-input" :placeholder="$t('configuration.themes.export.placeholder_version')" />
            </div>
            <div class="export-field">
              <span class="export-label">{{ $t('configuration.themes.export.label_web') }}</span>
              <input v-model="exportHomePage" class="export-input" :placeholder="$t('configuration.themes.export.placeholder_web')" />
            </div>
            <div class="export-acts">
              <button class="theme-btn theme-btn--cancel" @click="showExportForm = false">{{ $t('configuration.themes.buttons.cancel') }}</button>
              <button class="theme-btn theme-btn--save" @click="doExport" :disabled="exporting">{{ exporting ? $t('configuration.themes.buttons.exporting') : $t('configuration.themes.buttons.export') }}</button>
            </div>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.themes.subsection_installed') }}</h3>

          <div v-if="loadingThemes" class="theme-empty">{{ $t('configuration.themes.loading') }}</div>
          <div v-else-if="!installedThemes.length" class="theme-empty">{{ $t('configuration.themes.empty_themes') }}</div>

          <div v-for="t in installedThemes" :key="t.name" class="theme-card" :class="{ expanded: selectedTheme === t }">
            <div class="theme-card-main" @click="toggleThemeDetail(t)">
              <img
                v-if="t.thumbnail"
                :src="`file:///${t.thumbnail}`"
                class="theme-thumb"
                alt=""
                @error="(e: any) => e.target.style.display = 'none'"
              />
              <div v-else class="theme-thumb theme-thumb--placeholder" />
              <div class="theme-info">
                <strong>{{ t.name }}</strong>
                <span v-if="t.author" class="theme-meta">{{ $t('configuration.themes.theme_by', { author: t.author }) }}</span>
                <span v-if="t.version" class="theme-meta">{{ $t('configuration.themes.theme_version', { version: t.version }) }}</span>
              </div>
              <img :class="['theme-chevron', { open: selectedTheme === t }]" :src="'assets/svg/chevron-right.svg'" width="12" height="12" />
            </div>
            <Transition name="expand">
              <div v-if="selectedTheme === t" class="theme-detail">
                <div class="theme-detail-preview" @click="previewImg = `file:///${t.thumbnail}`" v-if="t.thumbnail">
                  <img :src="`file:///${t.thumbnail}`" class="theme-detail-img" alt="" @error="(e: any) => e.target.style.display = 'none'" />
                </div>
                <div class="theme-detail-gallery" v-if="themeGallery.length">
                  <img v-for="(g, gi) in themeGallery" :key="gi" :src="`file:///${g}`" class="theme-detail-gallery-img" alt="" loading="lazy" @click="previewImg = `file:///${g}`" @error="(e: any) => e.target.style.display = 'none'" />
                </div>
                <div class="theme-detail-meta">
                  <div class="theme-detail-row"><span class="td-label">{{ $t('configuration.themes.detail_name') }}</span><span class="td-val">{{ t.name }}</span></div>
                  <div v-if="t.author" class="theme-detail-row"><span class="td-label">{{ $t('configuration.themes.detail_author') }}</span><span class="td-val">{{ t.author }}</span></div>
                  <div v-if="t.version" class="theme-detail-row"><span class="td-label">{{ $t('configuration.themes.detail_version') }}</span><span class="td-val">{{ $t('configuration.themes.theme_version', { version: t.version }) }}</span></div>
                  <div v-if="t.homePage" class="theme-detail-row"><span class="td-label">{{ $t('configuration.themes.detail_web') }}</span><a class="td-val td-link" :href="t.homePage" target="_blank">{{ t.homePage }}</a></div>
                </div>
                <div class="theme-detail-acts">
                  <button class="theme-btn theme-btn--apply" @click="applyTheme(t.name)" :disabled="applying === t.name">
                    {{ applying === t.name ? $t('configuration.themes.buttons.applying') : $t('configuration.themes.buttons.apply') }}
                  </button>
                  <button class="theme-btn theme-btn--delete" @click="deleteTheme(t.name)">
                    <img :src="'assets/svg/trash.svg'" width="12" height="12" />
                    {{ $t('configuration.themes.buttons.delete') }}
                  </button>
                </div>
              </div>
            </Transition>
          </div>

          <div class="theme-default">
            <button class="theme-btn theme-btn--default" @click="resetTheme">{{ $t('configuration.themes.buttons.reset_default') }}</button>
          </div>
        </div>

        <div v-else-if="section === 'about'" key="about" class="section-content">
          <h2 class="content-title">{{ $t('configuration.about.title') }}</h2>

          <div class="about-card">
            <p class="about-desc">{{ $t('configuration.about.description') }}</p>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.about.studio') }}</h3>
          <div class="about-card">
            <p class="about-desc">{{ $t('configuration.about.studio_desc') }}</p>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.about.developer') }}</h3>
          <div class="about-card">
            <p class="about-desc">{{ $t('configuration.about.developer_desc') }}</p>
            <div class="about-links">
              <a href="https://github.com/Stepnicka012" target="_blank" class="about-link">
                <img src="../../assets/svg/logo-github.svg" width="14" height="14" />
                {{ $t('configuration.about.links.github_dev') }}
              </a>
            </div>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.about.tech_title') }}</h3>
          <div class="about-card">
            <p class="about-desc">{{ $t('configuration.about.tech_desc') }}</p>
          </div>

          <h3 class="subsection-title">{{ $t('configuration.about.links_title') }}</h3>
          <div class="about-links-grid">
            <a href="https://github.com/NovaStepStudio" target="_blank" class="about-link-btn">
              <img src="../../assets/svg/logo-github.svg" width="16" height="16" /> {{ $t('configuration.about.links.github') }}
            </a>
            <a href="https://discord.gg/37dYy9apwE" target="_blank" class="about-link-btn">
              <img src="../../assets/svg/logo-discord.svg" width="16" height="16" /> {{ $t('configuration.about.links.discord') }}
            </a>
            <a href="https://www.youtube.com/@Stepnicka012" target="_blank" class="about-link-btn">
              <img src="../../assets/svg/logo-youtube.svg" width="16" height="16" /> {{ $t('configuration.about.links.youtube') }}
            </a>
            <a href="https://www.instagram.com/stepnickast" target="_blank" class="about-link-btn">
              <img src="../../assets/svg/logo-instagram.svg" width="16" height="16" /> {{ $t('configuration.about.links.instagram') }}
            </a>
            <a href="https://whatsapp.com/channel/0029Vb8PVjB2UPBOewyceo0K" target="_blank" class="about-link-btn">
              <img src="../../assets/svg/globe.svg" width="16" height="16" /> {{ $t('configuration.about.links.whatsapp') }}
            </a>
            <a href="https://github.com/NovaStepStudio/StepLauncher" target="_blank" class="about-link-btn">
              <img src="../../assets/svg/logo-github.svg" width="16" height="16" /> {{ $t('configuration.about.links.source') }}
            </a>
            <a href="https://steplauncher.pages.dev" target="_blank" class="about-link-btn">
              <img src="../../assets/svg/globe.svg" width="16" height="16" /> {{ $t('configuration.about.links.website') }}
            </a>
          </div>
        </div>

      </transition>

      <div v-if="previewImg" class="gallery-overlay" @click.self="closePreview">
        <button class="gallery-overlay__close" @click="closePreview">
          <img :src="'assets/svg/x.svg'" width="18" height="18" />
        </button>
        <div class="gallery-overlay__viewport" @wheel.prevent="onGalleryZoom" @mousedown.prevent="startGalleryDrag" @mousemove.prevent="onGalleryDrag" @mouseup.prevent="endGalleryDrag" @mouseleave.prevent="endGalleryDrag">
          <img :src="previewImg" class="gallery-overlay__img" :style="galleryImgStyle" draggable="false" />
        </div>
        <div class="gallery-overlay__controls">
          <button class="gallery-overlay__ctrl" @click="galleryZoomIn">+</button>
          <span class="gallery-overlay__zoom-lbl">{{ Math.round(galleryZoom * 100) }}%</span>
          <button class="gallery-overlay__ctrl" @click="galleryZoomOut">−</button>
          <button class="gallery-overlay__ctrl gallery-overlay__ctrl--reset" @click="resetGalleryZoom">↺</button>
        </div>
      </div>

    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import ColorPicker from '../Components/ColorPicker/ColorPicker.vue'
import { useLanguage } from '../Composables/useLanguage'
import { getLangMeta } from '../i18n'

const emit = defineEmits<{ back: [] }>()
const { t, locale, locales, changeLanguage, getFlagIcon } = useLanguage()

const section = ref('launcher')

const sections = computed(() => [
  { key: 'launcher', label: t('configuration.sidebar.launcher'), icon: 'assets/svg/home.svg'},
  { key: 'minecraft', label: t('configuration.sidebar.minecraft'), icon: 'assets/svg/layers.svg'},
  { key: 'personalization', label: t('configuration.sidebar.personalization'), icon: 'assets/svg/settings.svg'},
  { key: 'themes', label: t('configuration.sidebar.themes'), icon: 'assets/svg/palette.svg'},
  { key: 'about', label: t('configuration.sidebar.about'), icon: '../../assets/svg/info.svg'},
])

const config = reactive({
  launcher: {
    autoStartMinecraft: true,
    hideOnLaunch: true,
    showConsole: false,
    showNews: true,
    filters: true,
    blur: true,
    hardwareAcceleration: true,
    shadows: true,
    discordRpc: true,
  },
  minecraft: {
    useRecommendedJava: true,
    maxConsoleEvents: 100,
    showNotificationOnLaunch: true,
    cleanBeforeLaunch: true,
    javaPath: '',
    fullscreen: false,
    windowWidth: 854,
    windowHeight: 480,
    maxRam: 4096,
    minRam: 2048,
    gcPreset: 'g1gc_optimized',
    gpuPreference: 'auto',
    jvmArgs: '-XX:+AlwaysPreTouch\n-XX:+DisableExplicitGC',
    gameArgs: '',
  },
  personalization: {
    titleBarColor: '#111',
    appBackground: "url('assets/background/RRE36/19.webp')",
    fontPrimary: 'Lexend',
    fontSecondary: 'Inter',
    accentColor: '#5cd0e7',
    modalAccent: '#5cd0e7',
    modalBackground: 'rgba(18, 18, 30, 0.97)',
    sidebarBackground: 'rgba(8,8,16,0.65)',
    panelBackground: 'rgba(8,8,16,0.55)',
    notificationErrorColor: '#e57373',
    notificationWarnColor: '#ffd54f',
    notificationSuccessColor: '#81c784',
    sidebarButtonColor: 'rgba(255,255,255,0.06)',
    macOSTitlebar: false,
    showIcon: true,
    invertPosition: false,
    sidebarPosition: 'left',
  },
})

const totalRam = ref(0)
const maxRam = ref(16384)
const bgPreset = ref("url('../assets/background/RRE36/19.webp')")

async function loadConfig() {
  try {
    const [c, sys, info] = await Promise.all([
      window.ConfigManager.Get(),
      window.ConfigManager.GetSystemInfo(),
      window.NovaCoreManager.GetInfo().catch(() => null),
    ])
    Object.assign(config.launcher, c.launcher)
    Object.assign(config.minecraft, c.minecraft)
    Object.assign(config.personalization, c.personalization)
    if (!config.minecraft.javaPath && info?.javaPath) {
      config.minecraft.javaPath = info.javaPath
    }
    if (config.minecraft.minRam > config.minecraft.maxRam) {
      config.minecraft.minRam = config.minecraft.maxRam
      window.ConfigManager.UpdateMinecraft({ minRam: config.minecraft.minRam }).catch(() => {})
    }
    bgPreset.value = c.personalization.appBackground
    totalRam.value = sys.totalRam || 0
    const totalGB = Math.floor((sys.totalRam || 0) / 1024 / 1024 / 1024)
    maxRam.value = Math.max(512, (totalGB || 4) * 1024)
    if (isNaN(maxRam.value)) maxRam.value = 16384
    if (config.minecraft.maxRam > maxRam.value) {
      config.minecraft.maxRam = maxRam.value
      window.ConfigManager.UpdateMinecraft({ maxRam: maxRam.value }).catch(() => {})
    }
    applyPersonalization()
    applyLauncherSetting()
  } catch {}
}

const recommendedRam = computed(() => Math.min(
  maxRam.value,
  Math.max(2048, Math.floor(totalRam.value / 1024 / 1024 / 1024 / 2) * 1024)
))

function applyLauncherSetting() {
  const l = config.launcher
  const root = document.documentElement
  root.classList.toggle('blur-disabled', !l.blur)
  root.classList.toggle('filters-disabled', !l.filters)
  root.classList.toggle('shadows-disabled', !l.shadows)
  root.classList.toggle('hide-console', !l.showConsole)
  root.classList.toggle('hide-news', !l.showNews)
}

async function saveLauncher(key: string) {
  try {
    await window.ConfigManager.UpdateLauncher({ [key]: config.launcher[key as keyof typeof config.launcher] })
  } catch {}
  applyLauncherSetting()
}

async function openLauncherFolder() {
  const info = await window.NovaCoreManager.GetInfo()
  await window.ElectronAPI.OpenPath(info.baseDir)
}

async function saveMinecraft(key: string) {
  if (key === 'maxRam' && config.minecraft.minRam > config.minecraft.maxRam) {
    config.minecraft.minRam = config.minecraft.maxRam
    window.ConfigManager.UpdateMinecraft({ minRam: config.minecraft.minRam }).catch(() => {})
  }
  try {
    await window.ConfigManager.UpdateMinecraft({ [key]: config.minecraft[key as keyof typeof config.minecraft] })
  } catch {}
}

async function savePersonalization(key: string) {
  try {
    await window.ConfigManager.UpdatePersonalization({ [key]: config.personalization[key as keyof typeof config.personalization] })
  } catch {}
  applyPersonalization()
}

function applyPersonalization() {
  const p = config.personalization
  document.documentElement.style.setProperty('--background-title-bar', p.titleBarColor)
  document.body.style.backgroundImage = p.appBackground
  document.documentElement.style.setProperty('--font-family-primary', `'${p.fontPrimary}'`)
  document.documentElement.style.setProperty('--font-family-secundary', `'${p.fontSecondary}'`)
  document.documentElement.style.setProperty('--accent-color', p.accentColor)
  document.documentElement.style.setProperty('--modal-accent', p.modalAccent || p.accentColor)
  document.documentElement.style.setProperty('--modal-bg', p.modalBackground || 'rgba(18, 18, 30, 0.97)')
  document.documentElement.style.setProperty('--sidebar-bg', p.sidebarBackground)
  document.documentElement.style.setProperty('--panel-bg', p.panelBackground)
  document.documentElement.style.setProperty('--notif-error-text', p.notificationErrorColor)
  document.documentElement.style.setProperty('--notif-warn-text', p.notificationWarnColor)
  document.documentElement.style.setProperty('--notif-success-text', p.notificationSuccessColor)
  document.documentElement.style.setProperty('--notif-error-bg', p.notificationErrorColor + '33')
  document.documentElement.style.setProperty('--notif-warn-bg', p.notificationWarnColor + '33')
  document.documentElement.style.setProperty('--notif-success-bg', p.notificationSuccessColor + '33')
  document.documentElement.style.setProperty('--notif-error-border', p.notificationErrorColor + '4d')
  document.documentElement.style.setProperty('--notif-warn-border', p.notificationWarnColor + '4d')
  document.documentElement.style.setProperty('--notif-success-border', p.notificationSuccessColor + '4d')
  document.documentElement.style.setProperty('--sidebar-btn-color', p.sidebarButtonColor)
  document.documentElement.style.setProperty('--sidebar-position', p.sidebarPosition || 'left')
  document.documentElement.classList.toggle('macos-titlebar', p.macOSTitlebar)
  document.documentElement.classList.toggle('hide-icon', !p.showIcon)
  document.documentElement.classList.toggle('invert-position', p.invertPosition)
  window.dispatchEvent(new CustomEvent('sidebar-position-changed', { detail: p.sidebarPosition || 'left' }))
}

async function browseJava() {
  const result = await window.ElectronAPI.OpenFileDialog({
    filters: [
      { name: t('configuration.themes.file_filters.java'), extensions: ['exe', 'bat', 'cmd', ''] },
      { name: t('configuration.themes.file_filters.all'), extensions: ['*'] },
    ],
  })
  if (!result.canceled && result.filePath) {
    config.minecraft.javaPath = result.filePath
    saveMinecraft('javaPath')
  }
}

function onBgPreset() {
  if (bgPreset.value === '__custom__') {
    browseBackground()
    return
  }
  config.personalization.appBackground = bgPreset.value
  savePersonalization('appBackground')
}

async function browseBackground() {
  const result = await window.ElectronAPI.OpenFileDialog({
    filters: [
      { name: t('configuration.themes.file_filters.images'), extensions: ['png', 'jpg', 'jpeg', 'webp', 'bmp'] },
      { name: t('configuration.themes.file_filters.all'), extensions: ['*'] },
    ],
  })
  if (!result.canceled && result.filePath) {
    const val = `url('file:///${result.filePath.replace(/\\/g, '/')}')`
    config.personalization.appBackground = val
    bgPreset.value = val
    savePersonalization('appBackground')
  } else {
    bgPreset.value = config.personalization.appBackground
  }
}

function formatRam(mb: number) {
  if (mb >= 1024) return `${+(mb / 1024).toFixed(1)} GB`
  return `${mb} MB`
}

const showExportForm = ref(false)
const galleryZoom = ref(1)
const galleryPanX = ref(0)
const galleryPanY = ref(0)
const galleryDragging = ref(false)
const galleryDragStart = ref({ x: 0, y: 0 })
const ZOOM_STEP = 0.25
const ZOOM_MIN = 0.5
const ZOOM_MAX = 5

const galleryImgStyle = computed(() => ({
  transform: `translate(${galleryPanX.value}px, ${galleryPanY.value}px) scale(${galleryZoom.value})`,
  cursor: galleryZoom.value > 1 ? 'grab' : 'default',
}))

function closePreview() {
  previewImg.value = ''
  resetGalleryZoom()
}

function resetGalleryZoom() {
  galleryZoom.value = 1
  galleryPanX.value = 0
  galleryPanY.value = 0
}

function galleryZoomIn() {
  galleryZoom.value = Math.min(galleryZoom.value + ZOOM_STEP, ZOOM_MAX)
}

function galleryZoomOut() {
  galleryZoom.value = Math.max(galleryZoom.value - ZOOM_STEP, ZOOM_MIN)
}

function onGalleryZoom(e: WheelEvent) {
  const delta = e.deltaY > 0 ? -ZOOM_STEP : ZOOM_STEP
  galleryZoom.value = Math.max(ZOOM_MIN, Math.min(ZOOM_MAX, galleryZoom.value + delta))
}

function startGalleryDrag(e: MouseEvent) {
  if (galleryZoom.value <= 1) return
  galleryDragging.value = true
  galleryDragStart.value = { x: e.clientX - galleryPanX.value, y: e.clientY - galleryPanY.value }
}

function onGalleryDrag(e: MouseEvent) {
  if (!galleryDragging.value) return
  galleryPanX.value = e.clientX - galleryDragStart.value.x
  galleryPanY.value = e.clientY - galleryDragStart.value.y
}

function endGalleryDrag() {
  galleryDragging.value = false
}

const exportName = ref('')
const exportAuthor = ref('')
const exportVersion = ref('1.0.0')
const exportHomePage = ref('')
const exporting = ref(false)
const installedThemes = ref<ThemeInfo[]>([])
const loadingThemes = ref(false)
const applying = ref('')
const selectedTheme = ref<ThemeInfo | null>(null)
const themeGallery = ref<string[]>([])
const previewText = computed(() => {
  return t('configuration.personalization.preview_text')
    .replace('{color}', `<span style="color: ${config.personalization.accentColor}">`)
    .replace('{/color}', '</span>')
})

const previewImg = ref('')

async function loadThemes() {
  loadingThemes.value = true
  try {
    const res = await window.ThemeManager.List()
    if (res.success && res.themes) installedThemes.value = res.themes
  } catch {}
  loadingThemes.value = false
}

async function doExport() {
  if (!exportName.value.trim()) return
  exporting.value = true
  try {
    await window.ThemeManager.Export({
      name: exportName.value.trim(),
      author: exportAuthor.value.trim(),
      homePage: exportHomePage.value.trim(),
      version: exportVersion.value.trim() || '1.0.0',
    })
    showExportForm.value = false
    exportName.value = ''
    exportAuthor.value = ''
    exportHomePage.value = ''
    exportVersion.value = '1.0.0'
    await loadThemes()
  } catch {}
  exporting.value = false
}

async function importTheme() {
  try {
    const res = await window.ThemeManager.Import()
    if (res.success) await loadThemes()
  } catch {}
}

async function applyTheme(name: string) {
  applying.value = name
  try {
    const res = await window.ThemeManager.GetConfig(name)
    if (res.success && res.config) {
      Object.assign(config.personalization, res.config)
      bgPreset.value = res.config.appBackground
      await window.ConfigManager.UpdatePersonalization(res.config)
      applyPersonalization()
    }
  } catch {}
  applying.value = ''
}

async function deleteTheme(name: string) {
  try {
    await window.ThemeManager.Delete(name)
    await loadThemes()
  } catch {}
}

async function resetTheme() {
  try {
    const c = await window.ConfigManager.Get()
    Object.assign(config.personalization, c.personalization)
    bgPreset.value = c.personalization.appBackground
    await window.ConfigManager.UpdatePersonalization(c.personalization)
    applyPersonalization()
  } catch {}
}

async function toggleThemeDetail(t: ThemeInfo) {
  if (selectedTheme.value === t) {
    selectedTheme.value = null
    themeGallery.value = []
    return
  }
  selectedTheme.value = t
  themeGallery.value = []
  try {
    const res = await window.ThemeManager.GetGallery(t.name)
    if (res.success && res.gallery) themeGallery.value = res.gallery
  } catch {}
}

onMounted(() => {
  loadConfig()
})

let unsubPos: (() => void) | null = null

onMounted(() => {
  const handler = (e: Event) => {
    config.personalization.sidebarPosition = (e as CustomEvent).detail
  }
  window.addEventListener('sidebar-position-changed', handler as EventListener)
  unsubPos = () => window.removeEventListener('sidebar-position-changed', handler as EventListener)
})

onUnmounted(() => {
  if (unsubPos) unsubPos()
})

watch(section, (s) => {
  if (s === 'themes') loadThemes()
})
</script>

<style scoped lang="scss" src="../Styles/Panels/ConfigurationPanel.scss"></style>
