<template>
  <div class="instances-panel" @keydown.escape="emit('close')" tabindex="0">
    <button class="panel-close" @click="emit('close')" :title="$t('instances.tooltips.close')">
      <img :src="'assets/svg/x.svg'" width="16" height="16" />
    </button>
    <div class="instances-sidebar">
      <button class="sidebar-add" @click="showCreate">+</button>

      <div class="sidebar-filters">
        <button :class="['sf-btn', { active: filterMode === 'all' }]" @click="filterMode = 'all'">{{ $t('instances.sidebar.filters.all') }}</button>
        <button :class="['sf-btn', { active: filterMode === 'favorites' }]" @click="filterMode = 'favorites'">
          <img :src="'assets/svg/star-filled.svg'" width="10" height="10" />
          {{ $t('instances.sidebar.filters.favorites') }}
        </button>
        <button :class="['sf-btn', { active: filterMode === 'pinned' }]" @click="filterMode = 'pinned'">
          <img :src="'assets/svg/pin-filled.svg'" width="10" height="10" />
          {{ $t('instances.sidebar.filters.pinned') }}
        </button>
      </div>

      <div v-if="sortedInstances.length" class="sidebar-list">
        <div
          v-for="inst in sortedInstances" :key="inst.id"
          :class="['sidebar-item', { active: selected?.id === inst.id }]"
          @click="selectInstance(inst)"
        >
          <div class="si-av">
            <img v-if="instIcon(inst)" :src="fileUrl(instIcon(inst))" class="si-img" />
            <span v-else>{{ displayName(inst).charAt(0).toUpperCase() }}</span>
          </div>
          <div class="si-info">
            <strong>{{ displayName(inst) }}</strong>
            <span class="si-meta">
              <span class="si-version">{{ versionLabel(inst) }}</span>
              <span class="si-tags">
                <span v-if="isFavorite(inst)" class="si-tag si-tag-fav" :title="$t('instances.tooltips.favorite')">
                  <img :src="'assets/svg/star-filled.svg'" width="8" height="8" />
                </span>
                <span v-if="isPinned(inst)" class="si-tag si-tag-pin" :title="$t('instances.tooltips.pinned')">
                  <img :src="'assets/svg/pin-filled.svg'" width="8" height="8" />
                </span>
              </span>
            </span>
          </div>
          <div class="si-actions">
            <button class="si-btn si-fav" :class="{ active: isFavorite(inst) }" @click.stop="toggleFavorite(inst)" :title="$t('instances.tooltips.favorite')">
              <img :src="isFavorite(inst) ? 'assets/svg/star-filled.svg' : 'assets/svg/star.svg'" width="11" height="11" />
            </button>
            <button class="si-btn si-pin" :class="{ active: isPinned(inst) }" @click.stop="togglePinned(inst)" :title="$t('instances.tooltips.pinned')">
              <img :src="isPinned(inst) ? 'assets/svg/pin-filled.svg' : 'assets/svg/pin.svg'" width="11" height="11" />
            </button>
          </div>
        </div>
      </div>
      <div v-else class="sidebar-empty">{{ $t('instances.sidebar.empty') }}</div>
    </div>

    <div class="instances-main">
      <div v-if="appInstallState.active && view !== 'create'" class="inst-progress-banner">
        <div class="ipb-header">
          <div class="ipb-spinner" :class="{ done: appInstallState.percent >= 100 && !appInstallState.error }" />
          <span class="ipb-name">{{ appInstallState.instanceName }}</span>
          <span class="ipb-pct">{{ appInstallState.percent }}%</span>
        </div>
        <div class="ipb-track">
          <div class="ipb-fill" :style="{ width: appInstallState.percent + '%' }"></div>
        </div>
        <div class="ipb-status">{{ appInstallState.error || appInstallState.statusText }}</div>
      </div>
      <div v-if="!sortedInstances.length && !selected && view === 'list'" class="empty-state">
        <img :src="'assets/svg/monitor.svg'" width="48" height="48" style="opacity:0.12" />
        <h3>{{ $t('instances.empty.title') }}</h3>
        <p>{{ $t('instances.empty.desc') }}</p>
        <button class="btn-accent" @click="showCreate">{{ $t('instances.buttons.create') }}</button>
      </div>

      <div v-if="sortedInstances.length && view === 'list'" class="instances-grid">
        <div v-for="inst in sortedInstances" :key="inst.id" class="instance-card" @click="selectInstance(inst)">
          <div v-if="instanceHero(inst)" class="ic-hero" :style="{ backgroundImage: `url('${fileUrl(instanceHero(inst))}')` }"></div>
          <div v-else class="ic-hero ic-hero-empty"></div>
          <div class="ic-body">
            <div class="ic-top">
              <img v-if="instIcon(inst)" :src="fileUrl(instIcon(inst))" class="ic-icon" />
              <span v-else class="ic-icon ic-icon-letter">{{ displayName(inst).charAt(0).toUpperCase() }}</span>
              <div class="ic-info">
                <h3 class="ic-title">{{ displayName(inst) }}</h3>
                <p v-if="description(inst)" class="ic-desc">{{ description(inst) }}</p>
                <span class="ic-version">{{ versionLabel(inst) }}</span>
              </div>
            </div>
            <div class="ic-actions">
              <button class="ic-play" @click.stop="quickLaunch(inst)">
                <img :src="'assets/svg/play.svg'" width="14" height="14" />
                {{ $t('instances.buttons.play') }}
              </button>
              <button class="ic-edit" @click.stop="quickEdit(inst)">
                <img :src="'assets/svg/edit.svg'" width="12" height="12" />
                {{ $t('instances.buttons.edit') }}
              </button>
              <div class="ic-tags">
                <button :class="['ic-tag', { active: isFavorite(inst) }]" @click.stop="toggleFavorite(inst)" :title="$t('instances.tooltips.favorite')">
                  <img :src="isFavorite(inst) ? 'assets/svg/star-filled.svg' : 'assets/svg/star.svg'" width="12" height="12" />
                </button>
                <button :class="['ic-tag', { active: isPinned(inst) }]" @click.stop="togglePinned(inst)" :title="$t('instances.tooltips.pinned')">
                  <img :src="isPinned(inst) ? 'assets/svg/pin-filled.svg' : 'assets/svg/pin.svg'" width="12" height="12" />
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="view === 'create'" class="create-view">
        <div class="cv-header">
          <button class="btn-icon" @click="showList">
            <img :src="'assets/svg/arrow-left.svg'" width="14" height="14" />
          </button>
          <span>{{ $t('instances.create.title') }}</span>
        </div>

        <div v-if="!installing" class="cv-body">
          <div class="cv-left">
            <div class="field">
              <label>{{ $t('instances.create.fields.name') }}</label>
              <input v-model="createName" :placeholder="$t('instances.create.fields.name_placeholder')" class="cv-input" />
            </div>
            <div class="field">
              <label>{{ $t('instances.create.fields.description') }}</label>
              <textarea v-model="createDescription" :placeholder="$t('instances.create.fields.description_placeholder')" class="cv-input" rows="2" />
            </div>
            <div class="field">
              <label>{{ $t('instances.edit.fields.icon') }}</label>
              <div class="file-row">
                <input v-model="createIcon" :placeholder="$t('instances.edit.fields.icon_placeholder')" class="cv-input file-input" />
                <button class="btn-file" @click="pickCreateIcon">{{ $t('instances.buttons.browse') }}</button>
              </div>
            </div>
            <div class="field">
              <label>{{ $t('instances.edit.fields.hero') }}</label>
              <div class="file-row">
                <input v-model="createHero" :placeholder="$t('instances.edit.fields.hero_placeholder')" class="cv-input file-input" />
                <button class="btn-file" @click="pickCreateHero">{{ $t('instances.buttons.browse') }}</button>
              </div>
            </div>
            <div class="field">
              <label>{{ $t('instances.create.fields.version') }}</label>
              <div class="search-box">
                <img class="search-icon" :src="'assets/svg/search.svg'" width="12" height="12" />
                <input v-model="searchQuery" :placeholder="$t('instances.create.fields.search_placeholder')" class="cv-input search-input" />
              </div>
            </div>
            <div class="version-tabs">
              <button v-for="t in versionTabs" :key="t.key" :class="['vt', { active: versionTab === t.key }]" @click="versionTab = t.key as any">{{ t.label }}</button>
            </div>
            <div class="version-list">
              <button v-for="v in filteredVersions" :key="v.id" :class="['version-item', { selected: selectedVersion?.id === v.id }]" @click="selectedVersion = v">
                <div class="v-left">
                  <span class="v-name">{{ v.id }}</span>
                  <span :class="['v-type', typeClass(v.type) as any]">{{ typeLabel(v.type) }}</span>
                </div>
                <img v-if="selectedVersion?.id === v.id" :src="'assets/svg/check.svg'" width="14" height="14" />
              </button>
              <div v-if="filteredVersions.length === 0" class="no-results">{{ $t('instances.create.no_results') }}</div>
            </div>
          </div>
          <div class="cv-right">
            <div class="field">
              <label>{{ $t('instances.create.fields.modloader') }}</label>
            </div>
            <div class="ml-list">
              <button v-for="ml in modloaderOptions" :key="ml.value" :class="['ml-item', { active: selectedModloader === ml.value }]" @click="selectedModloader = ml.value">
                <img :src="ml.icon" alt="" class="ml-icon" />
                <span class="ml-name">{{ ml.label }}</span>
                <img v-if="selectedModloader === ml.value" :src="'assets/svg/check.svg'" width="10" height="10" />
              </button>
            </div>
            <div class="summary-box">
              <div class="summary-line"><span class="sk">{{ $t('instances.create.summary.version') }}</span><span class="sv">{{ selectedVersion?.id || '—' }}</span></div>
              <div class="summary-line"><span class="sk">{{ $t('instances.create.summary.modloader') }}</span><span class="sv">{{ selectedModloader === 'none' ? $t('instances.create.options.vanilla') : selectedModloader }}</span></div>
              <div class="summary-line"><span class="sk">{{ $t('instances.create.summary.name') }}</span><span class="sv">{{ createName || '—' }}</span></div>
            </div>
            <button class="btn-download" :disabled="!canInstall" @click="doInstall">
              <img :src="'assets/svg/download.svg'" width="14" height="14" />
              {{ $t('instances.buttons.install') }}
            </button>
          </div>
        </div>

        <div v-if="installing" class="cv-progress">
          <div class="progress-info">
            <span class="progress-title">{{ installState.instanceName }}</span>
            <span class="progress-status">{{ installState.statusText || $t('instances.status.installing') }}</span>
          </div>
          <div :class="['progress-bar-track', { indeterminate: installState.processingModloader }]">
            <div v-if="!installState.processingModloader" class="progress-bar-fill" :style="{ width: installState.percent + '%' }"></div>
            <div v-else class="progress-bar-indet"></div>
          </div>
          <div v-if="installState.totalFiles > 0 && !installState.processingModloader" class="progress-details">
            <span>{{ $t('instances.status.files_progress', { completed: installState.completedFiles, total: installState.totalFiles }) }}</span>
            <span>{{ installState.downloadedMb.toFixed(1) }}/{{ installState.totalMb.toFixed(1) }} MB</span>
            <span>{{ installState.percent }}%</span>
          </div>
          <div v-if="installState.modules.length" class="section-list">
            <div v-for="m in installState.modules" :key="m.module" :class="['section-item', { 'active-section': m.status === 'downloading' }]">
              <div class="section-icon">
                <img v-if="m.status === 'completed'" :src="'assets/svg/check.svg'" width="14" height="14" />
                <img v-else-if="m.status === 'failed'" :src="'assets/svg/x.svg'" width="14" height="14" />
                <img v-else-if="m.status === 'downloading'" :src="'assets/svg/clock.svg'" width="14" height="14" />
                <img v-else :src="'assets/svg/stop-circle.svg'" width="14" height="14" style="opacity:0.3" />
              </div>
              <div class="section-info"><span class="section-name">{{ moduleLabel(m.module) }}</span></div>
              <span :class="['section-status', 'status-' + m.status]">{{ statusLabel(m.status) }}</span>
            </div>
          </div>
          <button v-if="installState.logs.length" class="logs-toggle" @click="showLogs = !showLogs">
            {{ $t(showLogs ? 'instances.buttons.hide_logs' : 'instances.buttons.show_logs', { count: installState.logs.length }) }}
          </button>
          <div v-if="showLogs && installState.logs.length" class="progress-logs">
            <div v-for="(l, i) in installState.logs" :key="i" class="log-line">{{ l }}</div>
          </div>
          <button v-if="installDone" class="btn-accent" @click="finishInstall">{{ $t('instances.buttons.accept') }}</button>
        </div>
      </div>

      <div v-if="selected && view === 'detail'" class="detail-view">
        <div v-if="cfgFrontendHero" class="dt-hero-wrap">
          <div class="dt-hero" :style="{ backgroundImage: `url('${fileUrl(cfgFrontendHero)}')` }"></div>
          <button class="dt-back" @click="showList">
            <img :src="'assets/svg/arrow-left.svg'" width="16" height="16" />
          </button>
        </div>
        <button v-else class="dt-back dt-back--standalone" @click="showList">
          <img :src="'assets/svg/arrow-left.svg'" width="16" height="16" />
        </button>

        <div class="dt-header">
          <img v-if="cfgFrontendIcon" :src="fileUrl(cfgFrontendIcon)" class="dt-header-icon" />
          <div class="dt-header-text">
            <h2 class="dt-header-name">
              {{ displayName(selected) }}
              <span class="dt-header-badges">
                <button :class="['hdr-badge', { active: isFavorite(selected) }]" @click="toggleFavorite(selected)" :title="$t('instances.tooltips.favorite')">
                  <img :src="isFavorite(selected) ? 'assets/svg/star-filled.svg' : 'assets/svg/star.svg'" width="12" height="12" />
                </button>
                <button :class="['hdr-badge', { active: isPinned(selected) }]" @click="togglePinned(selected)" :title="$t('instances.tooltips.pinned')">
                  <img :src="isPinned(selected) ? 'assets/svg/pin-filled.svg' : 'assets/svg/pin.svg'" width="12" height="12" />
                </button>
              </span>
            </h2>
            <p v-if="description(selected)" class="dt-header-desc">{{ description(selected) }}</p>
          </div>
        </div>

        <div v-if="!editing" class="dt-acts">
          <button v-if="isRunning" class="btn-play btn-kill" @click="doKill">
            <img :src="'assets/svg/stop.svg'" width="16" height="16" />
            {{ $t('instances.buttons.force_kill') }}
          </button>
          <button v-else class="btn-play" @click="doLaunch" :disabled="launching">
            <img :src="'assets/svg/play.svg'" width="16" height="16" />
            {{ launching ? $t('instances.buttons.launching') : $t('instances.buttons.play') }}
          </button>
          <button class="btn-ghost" @click="startEdit" :disabled="launching">{{ $t('instances.buttons.edit') }}</button>
          <button class="btn-danger" @click="doDelete" :disabled="launching">{{ $t('instances.buttons.delete') }}</button>
        </div>

        <div v-if="!editing" class="dt-tabs">
          <button v-for="t in tabs" :key="t.key" :class="['dt-tab', { active: activeTab === t.key }]" @click="activeTab = t.key">{{ t.label }}</button>
        </div>

        <div v-if="!editing && activeTab === 'info'" class="dt-tab-content">
          <div class="dt-info-grid">
            <div class="dt-info-item"><span class="dti-lbl">{{ $t('instances.detail.version') }}</span><span class="dti-val">{{ lastVersion(selected) }}</span></div>
            <div class="dt-info-item"><span class="dti-lbl">{{ $t('instances.detail.modloader') }}</span><span class="dti-val">{{ modLoader(selected) === 'vanilla' ? $t('instances.detail.modloader_none') : modLoader(selected) }}</span></div>
            <div class="dt-info-item"><span class="dti-lbl">{{ $t('instances.detail.ram') }}</span><span class="dti-val">{{ ramLabel(selected) }}</span></div>
            <div class="dt-info-item"><span class="dti-lbl">{{ $t('instances.detail.gc') }}</span><span class="dti-val">{{ gcLabel(selected.config?.configInstance?.gcPreset) }}</span></div>
            <div class="dt-info-item"><span class="dti-lbl">{{ $t('instances.detail.gpu') }}</span><span class="dti-val">{{ gpuLabel(selected.config?.configInstance?.gpuPreference) }}</span></div>
            <div class="dt-info-item"><span class="dti-lbl">{{ $t('instances.detail.window') }}</span><span class="dti-val">{{ selected.config?.configInstance?.window?.width ?? 854 }}x{{ selected.config?.configInstance?.window?.height ?? 480 }}</span></div>
            <div class="dt-info-item"><span class="dti-lbl">{{ $t('instances.detail.hw_accel') }}</span><span class="dti-val">{{ selected.config?.configInstance?.hardwareAccel ? $t('instances.detail.yes') : $t('instances.detail.no') }}</span></div>
            <div class="dt-info-item"><span class="dti-lbl">{{ $t('instances.detail.playtime') }}</span><span class="dti-val">{{ playTime(selected) }}</span></div>
            <div v-if="lastPlayedAt(selected)" class="dt-info-item"><span class="dti-lbl">{{ $t('instances.detail.last_played') }}</span><span class="dti-val">{{ fmtLastPlayed(lastPlayedAt(selected)) }}</span></div>
            <div class="dt-info-item dt-info-path"><span class="dti-lbl">{{ $t('instances.detail.path') }}</span><span class="dti-val">{{ selected.dir }} <button class="btn-path-open" @click.stop="openInstanceFolder" :title="$t('instances.buttons.open_folder')"><img :src="'assets/svg/folder.svg'" width="12" height="12" /></button></span></div>
          </div>
        </div>

        <div v-if="!editing && activeTab === 'mods'" class="dt-tab-content">
          <div class="dt-mods">
            <div class="dt-mods-header">
              <span class="dt-mods-title">{{ $t('instances.mods_section.title') }}</span>
              <span class="dt-mods-path">{{ selected.dir }}/mods/</span>
            </div>
            <div class="dt-mods-search">
              <input v-model="modQuery" class="cv-input" :placeholder="$t('instances.mods_section.search_placeholder')" @input="onModSearch" />
            </div>
            <div v-if="modLoading" class="dt-empty" style="padding: 2rem 0;">
              <div class="spinner" />
            </div>
            <div v-else-if="modError" class="dt-empty" style="padding: 2rem 0; color: #e57373;">{{ modError }}</div>
            <div v-else-if="modResults.length" class="dt-mods-results">
              <div v-for="m in modResults" :key="m.project_id" class="dt-mod-item" @click="openModDetail(m)">
                <img v-if="m.icon_url" :src="m.icon_url" class="dt-mod-icon" />
                <div class="dt-mod-info">
                  <strong>{{ m.title }}</strong>
                  <span class="dt-mod-desc">{{ m.description }}</span>
                  <span class="dt-mod-meta">{{ $t('instances.mods_section.metadata_format', { author: m.author, downloads: formatNumberMod(m.downloads) }) }}</span>
                </div>
                <button class="dt-mod-dl" @click.stop="downloadModToInstance(m)" :disabled="modDownloading === m.project_id">
                  {{ modDownloading === m.project_id ? $t('instances.mods_section.downloading') : $t('instances.mods_section.download') }}
                </button>
              </div>
              <button v-if="modHasMore && !modLoading" class="btn-ghost btn-sm" style="margin-top:0.5rem;align-self:center;" @click="loadMoreMods">{{ $t('instances.mods_section.loading_more') }}</button>
            </div>
            <div v-else class="dt-empty" style="padding: 2rem 0;">
              <span>{{ $t('instances.mods_section.desc_placeholder') }}</span>
            </div>
          </div>
        </div>

        <div v-if="!editing && activeTab === 'worlds'" class="dt-tab-content">
          <div v-if="worlds.length === 0" class="dt-empty">{{ $t('instances.mods_section.empty_worlds') }}</div>
          <div v-else class="dt-worlds">
            <div v-for="w in worlds" :key="w.folderName || w.name" class="dt-world">
              <img v-if="w.iconBase64" :src="w.iconBase64.startsWith('data:') ? w.iconBase64 : `data:image/png;base64,${w.iconBase64}`" class="dw-icon" />
              <div class="dw-info">
                <strong class="dw-name">{{ w.levelName || w.name }}</strong>
                <span class="dw-meta">{{ w.versionName || '?' }} · {{ formatDate(w.lastPlayed) }}</span>
              </div>
            </div>
          </div>
        </div>

        <div v-if="!editing && activeTab === 'screenshots'" class="dt-tab-content">
          <div v-if="screenshots.length === 0" class="dt-empty">{{ $t('instances.mods_section.empty_screenshots') }}</div>
          <div v-else class="dt-screenshots">
            <div v-for="s in screenshots" :key="s.name" class="dt-sshot">
              <img :src="fileUrl(s.path)" class="dss-img" @click="previewScreenshot = s" />
              <span class="dss-name">{{ s.name }}</span>
            </div>
          </div>
          <div v-if="previewScreenshot" class="dss-overlay" @click="previewScreenshot = null">
            <img :src="fileUrl(previewScreenshot.path)" class="dss-preview" />
          </div>
        </div>

        <div v-if="!editing && activeTab === 'console'" class="dt-tab-content">
          <div class="dt-console" ref="consoleEl">
            <div v-for="(l, i) in consoleLogs" :key="i" :class="['cl-line', 'cl-' + l.level.toLowerCase()]">
              <span class="cl-time">{{ l.time }}</span>
              <span class="cl-msg">{{ l.message }}</span>
            </div>
            <div v-if="consoleLogs.length === 0" class="dt-empty">{{ $t('instances.console.waiting') }}</div>
          </div>
          <div class="dt-console-acts">
            <button class="btn-ghost btn-sm" @click="clearConsole">{{ $t('instances.console.clear') }}</button>
            <button v-if="logFiles.length" class="btn-ghost btn-sm" @click="showLogFiles = !showLogFiles">{{ $t('instances.console.old_logs', { count: logFiles.length }) }}</button>
          </div>
          <div v-if="showLogFiles && logFiles.length" class="dt-logfiles">
            <button v-for="f in logFiles" :key="f.name" class="dt-lf-item" @click="loadLogFile(f.name)">
              <span class="lf-name">{{ f.name }}</span>
              <span class="lf-size">{{ $t('instances.console.logfile_kb', { size: (f.size / 1024).toFixed(1) }) }}</span>
            </button>
          </div>
        </div>

        <div v-if="editing" class="dt-edit">
          <div class="dt-edit-section">
            <h3 class="dt-edit-h">{{ $t('instances.edit.sections.personalization') }}</h3>
            <div class="dt-edit-grid">
              <div class="field"><label>{{ $t('instances.edit.fields.name') }}</label><input v-model="cfgFrontendName" /></div>
              <div class="field"><label>{{ $t('instances.edit.fields.description') }}</label><textarea v-model="cfgFrontendDesc" rows="2" /></div>
              <div class="field">
                <label>{{ $t('instances.edit.fields.icon') }}</label>
                <div class="file-row">
                  <input v-model="cfgFrontendIcon" :placeholder="$t('instances.edit.fields.icon_placeholder')" class="file-input" />
                  <button class="btn-file" @click="pickIcon">{{ $t('instances.buttons.browse') }}</button>
                </div>
              </div>
              <div class="field">
                <label>{{ $t('instances.edit.fields.hero') }}</label>
                <div class="file-row">
                  <input v-model="cfgFrontendHero" :placeholder="$t('instances.edit.fields.hero_placeholder')" class="file-input" />
                  <button class="btn-file" @click="pickHero">{{ $t('instances.buttons.browse') }}</button>
                </div>
              </div>
              <div class="field" v-if="availVersions.length">
                <label>{{ $t('instances.edit.fields.version') }}</label>
                <select v-model="cfgVersion">
                  <option v-for="v in availVersions" :key="v" :value="v">{{ v }}</option>
                </select>
              </div>
            </div>
          </div>

          <div class="dt-edit-section">
            <h3 class="dt-edit-h">{{ $t('instances.edit.sections.performance') }}</h3>
            <div class="dt-edit-grid dt-edit-grid2">
              <div class="field"><label>{{ $t('instances.edit.fields.ram_min') }}</label><input v-model.number="cfgMinRam" type="number" /></div>
              <div class="field"><label>{{ $t('instances.edit.fields.ram_max') }}</label><input v-model.number="cfgMaxRam" type="number" /></div>
              <div class="field"><label>{{ $t('instances.edit.fields.gc_preset') }}</label>
                <select v-model="cfgGcPreset">
                  <option value="none">{{ $t('instances.edit.gc_options.none') }}</option>
                  <option value="auto">{{ $t('instances.edit.gc_options.auto') }}</option>
                  <option value="disabled">{{ $t('instances.edit.gc_options.disabled') }}</option>
                  <option value="off">{{ $t('instances.edit.gc_options.off') }}</option>
                  <option value="g1gc_basic">{{ $t('instances.edit.gc_options.g1gc_basic') }}</option>
                  <option value="g1gc_optimized">{{ $t('instances.edit.gc_options.g1gc_optimized') }}</option>
                  <option value="zgc">{{ $t('instances.edit.gc_options.zgc') }}</option>
                  <option value="shenandoah">{{ $t('instances.edit.gc_options.shenandoah') }}</option>
                </select>
              </div>
              <div class="field"><label>{{ $t('instances.edit.fields.gpu_pref') }}</label>
                <select v-model="cfgGpuPref">
                  <option value="none">{{ $t('instances.edit.gpu_options.none') }}</option>
                  <option value="auto">{{ $t('instances.edit.gpu_options.auto') }}</option>
                  <option value="dgpu">{{ $t('instances.edit.gpu_options.dgpu') }}</option>
                  <option value="igpu">{{ $t('instances.edit.gpu_options.igpu') }}</option>
                </select>
              </div>
              <div class="field"><label>{{ $t('instances.edit.fields.hw_accel') }}</label>
                <select v-model="cfgHwAccel">
                  <option :value="false">{{ $t('instances.edit.hw_options.no') }}</option>
                  <option :value="true">{{ $t('instances.edit.hw_options.yes') }}</option>
                </select>
              </div>
              <div class="field"><label>{{ $t('instances.edit.fields.java_path') }}</label><input v-model="cfgJavaPath" placeholder="java" /></div>
            </div>
          </div>

          <div class="dt-edit-section">
            <h3 class="dt-edit-h">{{ $t('instances.edit.sections.window') }}</h3>
            <div class="dt-edit-grid dt-edit-grid2 dt-edit-grid3">
              <div class="field"><label>{{ $t('instances.edit.fields.width') }}</label><input v-model.number="cfgWinW" type="number" /></div>
              <div class="field"><label>{{ $t('instances.edit.fields.height') }}</label><input v-model.number="cfgWinH" type="number" /></div>
              <div class="field"><label>{{ $t('instances.edit.fields.fullscreen') }}</label>
                <select v-model="cfgWinFs">
                  <option :value="false">{{ $t('instances.edit.hw_options.no') }}</option>
                  <option :value="true">{{ $t('instances.edit.hw_options.yes') }}</option>
                </select>
              </div>
            </div>
          </div>

          <div class="dt-edit-section">
            <h3 class="dt-edit-h">{{ $t('instances.edit.sections.jvm') }}</h3>
            <div class="dt-edit-grid">
              <div class="field"><label>{{ $t('instances.edit.fields.extra_args') }}</label><textarea v-model="cfgJvmExtraArgs" rows="3" /></div>
              <div class="field"><label>{{ $t('instances.edit.fields.prepend_args') }}</label><textarea v-model="cfgJvmPrependArgs" rows="2" /></div>
            </div>
          </div>

          <div class="dt-edit-section">
            <h3 class="dt-edit-h">{{ $t('instances.edit.sections.game') }}</h3>
            <div class="dt-edit-grid">
              <div class="field"><label>{{ $t('instances.edit.fields.extra_game_args') }}</label><textarea v-model="cfgExtraGameArgs" rows="3" /></div>
            </div>
          </div>

          <div class="dt-edit-acts">
            <button class="btn-accent" @click="saveFullEdit">{{ $t('instances.edit.buttons.save') }}</button>
            <button class="btn-ghost" @click="editing = false; loadConfig(selected)">{{ $t('instances.edit.buttons.cancel') }}</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, reactive } from 'vue'
import { t } from '../i18n'
import { useAuth } from '../Composables/useAuth'
import type { InstallCardState } from '../Widgets/InstallProgressCard/InstallProgressCard.vue'

const { user } = useAuth()
const props = defineProps<{ appInstallState: InstallCardState }>()
const emit = defineEmits<{ close: []; installStarted: [{ name: string }] }>()

const view = ref<'list' | 'create' | 'detail'>('list')
const instances = ref<any[]>([])
const selected = ref<any | null>(null)
const editing = ref(false)
const showLogs = ref(false)
const launching = ref(false)
const runningLaunchId = ref('')
const runningPollTimer = ref<ReturnType<typeof setInterval> | null>(null)
const isRunning = ref(false)
const filterMode = ref<'all' | 'favorites' | 'pinned'>('all')

function instanceCF(inst: any) { return inst.config?.configInstance?.customFields || {} }
function isFavorite(inst: any) { return instanceCF(inst).favorite === true }
function isPinned(inst: any) { return instanceCF(inst).pinned === true }
function lastPlayedAt(inst: any) { return instanceCF(inst).lastPlayedAt || '' }

const sortedInstances = computed(() => {
  let list = [...instances.value]
  if (filterMode.value === 'favorites') list = list.filter(isFavorite)
  else if (filterMode.value === 'pinned') list = list.filter(isPinned)
  list.sort((a, b) => {
    const aPinned = isPinned(a) ? 1 : 0
    const bPinned = isPinned(b) ? 1 : 0
    if (aPinned !== bPinned) return bPinned - aPinned
    const aFav = isFavorite(a) ? 1 : 0
    const bFav = isFavorite(b) ? 1 : 0
    if (aFav !== bFav) return bFav - aFav
    const aTime = lastPlayedAt(a) || ''
    const bTime = lastPlayedAt(b) || ''
    if (aTime && bTime) return bTime.localeCompare(aTime)
    if (aTime) return -1
    if (bTime) return 1
    return displayName(a).localeCompare(displayName(b))
  })
  return list
})

const activeTab = ref('info')
const tabs = computed(() => {
  const base = [
    { key: 'info', label: t('instances.detail.tabs.info') },
    { key: 'worlds', label: t('instances.detail.tabs.worlds') },
    { key: 'screenshots', label: t('instances.detail.tabs.screenshots') },
    { key: 'console', label: t('instances.detail.tabs.console') },
  ]
  if (selected.value && modLoader(selected.value) !== 'vanilla') {
    base.splice(1, 0, { key: 'mods', label: t('instances.detail.tabs.mods') })
  }
  return base
})
const availVersions = computed(() => {
  const versions = selected.value?.metadata?.installedVersions
  if (!versions?.length) return []
  return versions.map((v: any) => v.mcVersion).filter(Boolean)
})
const worlds = ref<any[]>([])
const screenshots = ref<any[]>([])
const consoleLogs = ref<{ time: string; message: string; level: string }[]>([])
const logFiles = ref<{ name: string; size: number; mtime: string }[]>([])
const showLogFiles = ref(false)
const previewScreenshot = ref<any>(null)
const consoleEl = ref<HTMLElement | null>(null)
let consoleUnsub: (() => void) | null = null
const MAX_CONSOLE_LINES = 500

const cfgFrontendName = ref('')
const cfgFrontendDesc = ref('')
const cfgFrontendIcon = ref('')
const cfgFrontendHero = ref('')
const cfgMinRam = ref(512)
const cfgMaxRam = ref(2048)
const cfgGcPreset = ref('none')
const cfgGpuPref = ref('none')
const cfgHwAccel = ref(false)
const cfgJavaPath = ref('java')
const defaultJavaPath = ref('')
const cfgWinW = ref(854)
const cfgWinH = ref(480)
const cfgWinFs = ref(false)
const cfgJvmExtraArgs = ref('')
const cfgJvmPrependArgs = ref('')
const cfgExtraGameArgs = ref('')
const cfgVersion = ref('')

const launchStartTime = ref(0)
let playtimeTimer: ReturnType<typeof setInterval> | null = null
let launchExitUnsub: (() => void) | null = null

function loadConfig(inst: any) {
  const c = inst?.config
  cfgFrontendName.value = c?.instanceMetadata?.frontend?.name || inst?.id || ''
  cfgFrontendDesc.value = c?.instanceMetadata?.frontend?.description || ''
  cfgFrontendIcon.value = c?.instanceMetadata?.frontend?.icon || ''
  cfgFrontendHero.value = c?.instanceMetadata?.frontend?.hero || ''
  cfgMinRam.value = c?.configInstance?.minMemoryMb ?? 512
  cfgMaxRam.value = c?.configInstance?.maxMemoryMb ?? 2048
  cfgGcPreset.value = c?.configInstance?.gcPreset || 'none'
  cfgGpuPref.value = c?.configInstance?.gpuPreference || 'none'
  cfgHwAccel.value = c?.configInstance?.hardwareAccel ?? false
  cfgJavaPath.value = c?.configInstance?.javaPath || defaultJavaPath.value || 'java'
  cfgWinW.value = c?.configInstance?.window?.width ?? 854
  cfgWinH.value = c?.configInstance?.window?.height ?? 480
  cfgWinFs.value = c?.configInstance?.window?.fullscreen ?? false
  cfgJvmExtraArgs.value = (c?.configInstance?.jvm?.extraArgs || []).join('\n')
  cfgJvmPrependArgs.value = (c?.configInstance?.jvm?.prependArgs || []).join('\n')
  cfgExtraGameArgs.value = (c?.configInstance?.extraGameArgs || []).join('\n')
  cfgVersion.value = c?.configInstance?.customFields?.selectedVersion || lastVersion(inst)
}

const createName = ref('')
const createDescription = ref('')
const createIcon = ref('')
const createHero = ref('')
const searchQuery = ref('')
const versionTab = ref<'all' | 'release' | 'snapshot' | 'old'>('release')
const selectedVersion = ref<any | null>(null)
const selectedModloader = ref('none')
const allVersions = ref<any[]>([])
const loadingVersions = ref(true)
const installing = ref(false)

const installState = reactive({
  instanceName: '',
  percent: 0,
  completedFiles: 0,
  totalFiles: 0,
  downloadedMb: 0,
  totalMb: 0,
  statusText: '',
  modules: [] as any[],
  logs: [] as string[],
  processingModloader: false,
})

let installUnsub: (() => void) | null = null

const versionTabs = computed(() => [
  { key: 'all', label: t('instances.create.tabs.all') },
  { key: 'release', label: t('instances.create.tabs.release') },
  { key: 'snapshot', label: t('instances.create.tabs.snapshot') },
  { key: 'old', label: t('instances.create.tabs.old') },
])

const modloaderOptions = computed(() => [
  { value: 'none', label: t('instances.create.options.vanilla'), icon: 'assets/loaders/Grass.webp' },
  { value: 'fabric', label: t('instances.create.options.fabric'), icon: 'assets/loaders/fabric.png' },
  { value: 'forge', label: t('instances.create.options.forge'), icon: 'assets/loaders/forge.png' },
  { value: 'neoforge', label: t('instances.create.options.neoforge'), icon: 'assets/loaders/neoforge.png' },
  { value: 'quilt', label: t('instances.create.options.quilt'), icon: 'assets/loaders/quilt.png' },
])

const modQuery = ref('')
const modResults = ref<any[]>([])
const modLoading = ref(false)
const modError = ref('')
const modOffset = ref(0)
const modHasMore = ref(false)
const modDownloading = ref('')
let modSearchTimer: ReturnType<typeof setTimeout> | null = null

function formatNumberMod(n: number): string {
  if (!n) return '0'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'k'
  return String(n)
}

function onModSearch() {
  if (modSearchTimer) clearTimeout(modSearchTimer)
  modSearchTimer = setTimeout(() => searchMods(), 400)
}

async function searchMods() {
    modLoading.value = true
    modError.value = ''
    modOffset.value = 0
  try {
    const params = new URLSearchParams()
    if (modQuery.value.trim()) params.set('query', modQuery.value.trim())
    params.set('limit', '20')
    params.set('offset', '0')
    params.set('index', 'relevance')
    params.set('facets', JSON.stringify([['project_type:mod']]))
    const res = await fetch(`https://api.modrinth.com/v2/search?${params}`)
    if (!res.ok) throw new Error(`Error ${res.status}`)
    const data = await res.json()
    modResults.value = data.hits || []
    modHasMore.value = (data.total_hits || 0) > 20
  } catch (e: any) {
    modError.value = e?.message || t('instances.mods_section.search_error')
    modResults.value = []
  }
  modLoading.value = false
}

async function loadMoreMods() {
  modOffset.value += 20
  try {
    const params = new URLSearchParams()
    if (modQuery.value.trim()) params.set('query', modQuery.value.trim())
    params.set('limit', '20')
    params.set('offset', String(modOffset.value))
    params.set('index', 'relevance')
    params.set('facets', JSON.stringify([['project_type:mod']]))
    const res = await fetch(`https://api.modrinth.com/v2/search?${params}`)
    if (!res.ok) throw new Error()
    const data = await res.json()
    modResults.value = [...modResults.value, ...(data.hits || [])]
    modHasMore.value = (data.total_hits || 0) > modOffset.value + 20
  } catch {}
}

async function downloadModToInstance(m: any) {
  if (!selected.value) return
  try {
    const verRes = await fetch(`https://api.modrinth.com/v2/project/${m.project_id}/version`)
    if (!verRes.ok) return
    const versions = await verRes.json()
    const latest = versions?.[0]
    if (!latest?.files?.[0]) return
    const file = latest.files[0]
    const destPath = `${selected.value.dir}/mods/${file.filename}`
    modDownloading.value = m.project_id
    await window.ModsManager.DownloadToPath(file.url, destPath)
  } catch {}
  modDownloading.value = ''
}

async function openModDetail(m: any) {
  window.ElectronAPI.OpenExternal(`https://modrinth.com/mod/${m.project_id}`)
}

function moduleLabel(m: string) {
  const key = `instances.modules.${m}`
  const translated = t(key)
  return translated !== key ? translated : m
}
function statusLabel(s: string) {
  const key = `instances.states.${s}`
  const translated = t(key)
  return translated !== key ? translated : s
}
function typeLabel(type: string) {
  const map: Record<string, string> = {
    release: 'modals.download.types.release',
    snapshot: 'modals.download.types.snapshot',
    old_beta: 'modals.download.types.beta',
    old_alpha: 'modals.download.types.alpha',
  }
  return map[type] ? t(map[type]) : type
}
function typeClass(t: string) {
  const map: Record<string, string> = { release: 'release', snapshot: 'snapshot', old_beta: 'old', old_alpha: 'old' }
  return map[t] || ''
}

const filteredVersions = computed(() => {
  let list = allVersions.value
  const q = searchQuery.value.toLowerCase().trim()
  if (q) list = list.filter(v => v.id.toLowerCase().includes(q))
  if (versionTab.value === 'release') list = list.filter(v => v.type === 'release')
  else if (versionTab.value === 'snapshot') list = list.filter(v => v.type === 'snapshot')
  else if (versionTab.value === 'old') list = list.filter(v => v.type === 'old_beta' || v.type === 'old_alpha')
  return list
})

const canInstall = computed(() => !installing.value && !!selectedVersion.value && !!createName.value.trim())
const installFailed = computed(() => installState.statusText.startsWith('Error:'))
const installDone = computed(() => installFailed.value || (
  !installState.processingModloader &&
  installState.totalFiles > 0 &&
  installState.completedFiles >= installState.totalFiles
))

function displayName(inst: any) { return inst.config?.instanceMetadata?.frontend?.name || inst.id }
function versionLabel(inst: any) {
  const v = lastVersion(inst)
  const ml = modLoader(inst)
  return ml !== 'vanilla' ? `${v} · ${ml}` : v
}
function lastVersion(inst: any) {
  const versions = inst.metadata?.installedVersions
  if (versions?.length) {
    const ml = modLoader(inst)
    if (ml !== 'vanilla') {
      const mlVer = versions.find((v: any) => v.mcVersion && v.mcVersion.includes(ml))
      if (mlVer) return mlVer.mcVersion
    }
    return (versions.find((v: any) => v.lastPlayed) || versions[versions.length - 1])?.mcVersion || '?'
  }
  return '?'
}
function modLoader(inst: any) { return inst.config?.configInstance?.modLoader || 'vanilla' }
function ramLabel(inst: any) {
  const min = inst.config?.configInstance?.jvm?.minMemoryMb ?? inst.config?.configInstance?.minMemoryMb ?? 512
  const max = inst.config?.configInstance?.jvm?.maxMemoryMb ?? inst.config?.configInstance?.maxMemoryMb ?? 2048
  return `${min} MB - ${max} MB`
}
function description(inst: any) { return inst.config?.instanceMetadata?.frontend?.description || '' }
function playTime(inst: any) {
  const ms = inst.config?.instanceMetadata?.totalPlayTimeMs ?? 0
  const h = Math.floor(ms / 3600000)
  const m = Math.floor((ms % 3600000) / 60000)
  return h > 0 ? `${h}h ${m}m` : `${m}m`
}
function instIcon(inst: any) { return inst.config?.instanceMetadata?.frontend?.icon || null }
function instanceHero(inst: any) { return inst.config?.instanceMetadata?.frontend?.hero || null }
function fileUrl(path: string) {
  if (!path) return ''
  if (path.startsWith('http')) return path
  if (path.startsWith('file://')) return path
  path = path.replace(/\\/g, '/')
  return `file:///${encodeURI(path)}`
}
function gcLabel(v: string) {
  if (!v || v === 'none') return t('instances.detail.gc_label_none')
  const key = `instances.edit.gc_options.${v}`
  const translated = t(key)
  return translated !== key ? translated : v
}
function gpuLabel(v: string) {
  if (!v || v === 'none') return t('instances.detail.gpu_label_none')
  const key = `instances.edit.gpu_options.${v}`
  const translated = t(key)
  return translated !== key ? translated : v
}
function formatDate(ts: number) {
  if (!ts) return ''
  const d = new Date(ts)
  return d.toLocaleDateString('es', { day: 'numeric', month: 'short', year: 'numeric' })
}
function fmtLastPlayed(iso: string) {
  if (!iso) return ''
  const d = new Date(iso)
  return d.toLocaleDateString('es', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
}

async function loadWorlds() {
  if (!selected.value) return
  try {
    const apiWorlds = await window.NovaCoreManager.Worlds().catch(() => null)
    if (apiWorlds && apiWorlds.worlds && apiWorlds.worlds.length) {
      worlds.value = apiWorlds.worlds
      return
    }
  } catch {}
  try {
    const savesDir = `${selected.value.dir}/game/saves`
    const entries = await window.ElectronAPI.ReadDir(savesDir)
    if (entries) {
      worlds.value = entries.filter(e => e.isDirectory).map(e => ({
        name: e.name,
        folderName: e.name,
        levelName: e.name,
        lastPlayed: 0,
        versionName: '',
        iconBase64: null,
      }))
    }
  } catch {
    worlds.value = []
  }
}

async function loadScreenshots() {
  if (!selected.value) return
  try {
    const ssDir = `${selected.value.dir}/game/screenshots`
    const entries = await window.ElectronAPI.ReadDir(ssDir)
    if (!entries) {
      screenshots.value = []
      return
    }
    screenshots.value = entries
      .filter(e => e.isFile && /\.(png|jpg|jpeg|webp)$/i.test(e.name))
      .sort((a, b) => b.mtime.localeCompare(a.mtime))
      .map(e => ({ name: e.name, path: `${ssDir}/${e.name}`, mtime: e.mtime }))
  } catch {
    screenshots.value = []
  }
}

function subscribeConsole() {
  consoleUnsub?.()
  consoleUnsub = window.NovaCoreManager.OnEvent((ev) => {
    if (ev.event === 'game_log') {
      const d = ev.data as any
      const time = new Date().toLocaleTimeString('es', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
      consoleLogs.value.push({ time, message: d.message || d.line || '', level: d.level || 'INFO' })
      if (consoleLogs.value.length > MAX_CONSOLE_LINES) consoleLogs.value.splice(0, consoleLogs.value.length - MAX_CONSOLE_LINES)
      nextTick(() => { if (consoleEl.value) consoleEl.value.scrollTop = consoleEl.value.scrollHeight })
    }
  })
}

function clearConsole() {
  consoleLogs.value = []
}

async function loadLogFiles() {
  try {
    logFiles.value = await window.NovaCoreManager.GetLogFiles()
  } catch {
    logFiles.value = []
  }
}

async function loadLogFile(name: string) {
  try {
    const lines = await window.NovaCoreManager.ReadLogFile(name, 500)
    if (lines) {
      consoleLogs.value = lines.map(l => ({
        time: '', message: l, level: 'INFO',
      }))
      nextTick(() => { if (consoleEl.value) consoleEl.value.scrollTop = consoleEl.value.scrollHeight })
    }
  } catch {}
}

function showList() {
  view.value = 'list'
  selected.value = null
}

function showCreate() {
  view.value = 'create'
  selected.value = null
  createName.value = ''
  createDescription.value = ''
  createIcon.value = ''
  createHero.value = ''
  searchQuery.value = ''
  selectedVersion.value = null
  selectedModloader.value = 'none'
  versionTab.value = 'release'
}

async function toggleFavorite(inst: any) {
  const cf = { ...instanceCF(inst), favorite: !isFavorite(inst) }
  await window.InstancesManager.UpdateConfig(inst.id, { configInstance: { customFields: cf } })
  await loadInstances()
}

async function togglePinned(inst: any) {
  const cf = { ...instanceCF(inst), pinned: !isPinned(inst) }
  await window.InstancesManager.UpdateConfig(inst.id, { configInstance: { customFields: cf } })
  await loadInstances()
}

async function selectInstance(inst: any) {
  selected.value = inst
  view.value = 'detail'
  activeTab.value = 'info'
  loadConfig(inst)
  loadWorlds()
  loadScreenshots()
  subscribeConsole()
  loadLogFiles()
}

async function quickLaunch(inst: any) {
  selected.value = inst
  view.value = 'detail'
  activeTab.value = 'info'
  loadConfig(inst)
  loadWorlds()
  loadScreenshots()
  subscribeConsole()
  loadLogFiles()
  await nextTick()
  await doLaunch()
}

function quickEdit(inst: any) {
  selectInstance(inst)
  nextTick(() => startEdit())
}

async function loadInstances() {
  try {
    instances.value = await window.InstancesManager.List()
  } catch {
    instances.value = []
  }
}

async function loadVersions() {
  loadingVersions.value = true
  try {
    const vRes = await window.NovaCoreManager.Versions(200).catch(() => null)
    if (vRes) {
      const raw = (vRes as any).versions ?? vRes
      if (Array.isArray(raw)) {
        const seen = new Set<string>()
        allVersions.value = raw.filter((v: any) => {
          if (!v.id || seen.has(v.id)) return false
          seen.add(v.id)
          return true
        })
        const first = allVersions.value.find(v => v.type === 'release')
        if (first) selectedVersion.value = first
      }
    }
  } catch {
  } finally {
    loadingVersions.value = false
  }
}

async function doInstall() {
  if (!selectedVersion.value || !createName.value.trim()) return
  const name = createName.value.trim()
  const dirName = name.replace(/[^a-zA-Z0-9\-_.]/g, '_')
  const instancesDir = await window.InstancesManager.GetDir()
  const instanceDir = `${instancesDir}/${dirName}`
  const sharedDir = await window.InstancesManager.GetSharedDir()

  installing.value = true
  emit('installStarted', { name })
  installState.instanceName = name
  installState.percent = 0
  installState.completedFiles = 0
  installState.totalFiles = 0
  installState.downloadedMb = 0
  installState.totalMb = 0
  installState.statusText = t('instances.status.starting_install')
  installState.modules = []
  installState.logs = []

  try {
    await window.InstancesManager.UpdateConfig(dirName, {
      name,
      description: createDescription.value.trim(),
      icon: createIcon.value,
      hero: createHero.value,
      configInstance: {
        modLoader: selectedModloader.value,
        minMemoryMb: 4096,
        maxMemoryMb: 4096,
      }
    })

    installUnsub = window.NovaCoreManager.OnEvent((ev: any) => {
      const { event, data } = ev
      if (!data) return
      switch (event) {
        case 'install_step':
          if (data.step === 'modloader') {
            installState.processingModloader = true
            installState.statusText = t('instances.status.processing_modloader', { loader: data.loader || 'modloader' })
          } else {
            installState.statusText = data.step ?? installState.statusText
          }
          break
        case 'modloader_resolving':
          installState.processingModloader = true
          installState.statusText = t('instances.status.resolving_modloader', { loader: data.loader || 'modloader' })
          break
        case 'modloader_downloading':
          installState.statusText = t('instances.status.downloading_modloader', { loader: data.loader || 'modloader' })
          break
        case 'modloader_install_start':
          installState.statusText = t('instances.status.installing_modloader', { loader: data.loader || 'modloader' })
          break
        case 'modloader_processor_log':
          if (data.line && !data.line.startsWith('[')) {
            installState.logs.push(`[modloader] ${data.line}`)
            if (installState.logs.length > 200) installState.logs.splice(0, 100)
          }
          break
        case 'tasks_ready':
          installState.totalFiles = data.totalTasks ?? 0
          installState.totalMb = (data.totalBytes ?? 0) / 1048576
          installState.statusText = t('instances.status.preparing')
          break
        case 'session_started':
          installState.totalFiles = data.totalFiles ?? 0
          installState.totalMb = (data.totalBytes ?? 0) / 1048576
          break
        case 'session_progress':
          installState.percent = data.overallPercent ?? data.percent ?? 0
          installState.completedFiles = data.completedFiles ?? 0
          installState.downloadedMb = (data.downloadedBytes ?? data.totalBytes ?? 0) / 1048576
          installState.statusText = data.step ?? installState.statusText
          break
        case 'module_status':
          if (data.module && data.status) {
            const idx = installState.modules.findIndex((m: any) => m.module === data.module)
            if (idx >= 0) installState.modules[idx]!.status = data.status
            else installState.modules.push({ module: data.module, status: data.status })
          }
          break
        case 'modloader_installed':
        case 'install_completed':
          installState.processingModloader = false
          installState.percent = 100
          installState.completedFiles = installState.totalFiles
          installState.downloadedMb = installState.totalMb
          installState.statusText = t('instances.status.install_completed')
          break
        case 'session_completed':
          if (installState.totalFiles === 0) {
            installState.totalFiles = data.totalFiles ?? 0
            installState.totalMb = (data.downloadedBytes ?? 0) / 1048576
          }
          break
        case 'session_failed':
        case 'install_failed':
          installState.processingModloader = false
          installState.statusText = t('instances.status.error', { error: data.reason || data.error || t('instances.status.error_install') })
          installState.percent = 100
          break
      }
    })

    const installReq: any = {
      version: selectedVersion.value.id,
      instancePath: instanceDir,
      sharedPath: sharedDir,
      isInstance: true,
      launcher: { name: 'StepLauncher', version: '1.0.0' },
      download: { jvm: true },
    }
    if (selectedModloader.value !== 'none') {
      installReq.modloader = selectedModloader.value
    }
    const installResult = await window.NovaCoreManager.Install(installReq)
    if (!installResult.success) {
      installState.statusText = t('instances.status.error', { error: installResult.error || t('instances.status.error_install') })
      installState.percent = 100
      return
    }

    await window.InstancesManager.UpdateConfig(dirName, {
      name: name,
      description: createDescription.value.trim(),
      icon: createIcon.value,
      hero: createHero.value,
      configInstance: {
        modLoader: selectedModloader.value === 'none' ? 'vanilla' : selectedModloader.value,
      }
    })
  } catch (err: any) {
    installState.statusText = `Error: ${err?.message ?? '?'}`
  }
}

async function finishInstall() {
  if (installUnsub) {
    installUnsub()
    installUnsub = null
  }
  installing.value = false

  if (installFailed.value) {
    view.value = 'list'
    selected.value = null
    loadInstances()
    return
  }

  const cfg = await window.ConfigManager.Get().catch(() => null)
  const autoLaunch = cfg?.launcher?.autoStartMinecraft ?? false

  if (autoLaunch) {
    await loadInstances()
    const created = instances.value.find((i: any) =>
      i.config?.instanceMetadata?.frontend?.name === installState.instanceName ||
      i.id === installState.instanceName ||
      i.id.includes(installState.instanceName.replace(/[^a-zA-Z0-9\-_.]/g, '_'))
    )
    if (created) {
      view.value = 'detail'
      selected.value = created
      loadConfig(created)
      loadWorlds()
      loadScreenshots()
      subscribeConsole()
      loadLogFiles()
      await nextTick()
      doLaunch()
      return
    }
  }

  view.value = 'list'
  selected.value = null
  loadInstances()
}

function startEdit() {
  if (!selected.value) return
  loadConfig(selected.value)
  editing.value = true
}

async function pickCreateIcon() {
  const result = await window.ElectronAPI.OpenFileDialog({ filters: [{ name: 'Imágenes', extensions: ['png', 'jpg', 'jpeg', 'webp', 'svg'] }] })
  if (result && !result.canceled && result.filePath) createIcon.value = result.filePath
}

async function pickCreateHero() {
  const result = await window.ElectronAPI.OpenFileDialog({ filters: [{ name: 'Imágenes', extensions: ['png', 'jpg', 'jpeg', 'webp', 'svg'] }] })
  if (result && !result.canceled && result.filePath) createHero.value = result.filePath
}

async function pickIcon() {
  const result = await window.ElectronAPI.OpenFileDialog({ filters: [{ name: 'Imágenes', extensions: ['png', 'jpg', 'jpeg', 'webp', 'svg'] }] })
  if (result && !result.canceled && result.filePath) cfgFrontendIcon.value = result.filePath
}

async function pickHero() {
  const result = await window.ElectronAPI.OpenFileDialog({ filters: [{ name: 'Imágenes', extensions: ['png', 'jpg', 'jpeg', 'webp', 'svg'] }] })
  if (result && !result.canceled && result.filePath) cfgFrontendHero.value = result.filePath
}

async function saveFullEdit() {
  if (!selected.value) return
  const id = selected.value.id
  try {
    await window.InstancesManager.UpdateConfig(id, {
      name: cfgFrontendName.value.trim() || id,
      description: cfgFrontendDesc.value.trim(),
      icon: cfgFrontendIcon.value,
      hero: cfgFrontendHero.value,
      configInstance: {
        minMemoryMb: cfgMinRam.value,
        maxMemoryMb: cfgMaxRam.value,
        gcPreset: cfgGcPreset.value,
        gpuPreference: cfgGpuPref.value,
        hardwareAccel: cfgHwAccel.value,
        javaPath: cfgJavaPath.value === defaultJavaPath.value ? '' : cfgJavaPath.value,
        window: {
          width: cfgWinW.value,
          height: cfgWinH.value,
          fullscreen: cfgWinFs.value,
        },
        jvm: {
          extraArgs: cfgJvmExtraArgs.value.split('\n').map(s => s.trim()).filter(Boolean),
          prependArgs: cfgJvmPrependArgs.value.split('\n').map(s => s.trim()).filter(Boolean),
        },
        extraGameArgs: cfgExtraGameArgs.value.split('\n').map(s => s.trim()).filter(Boolean),
        customFields: { ...instanceCF(selected.value), selectedVersion: cfgVersion.value },
      }
    })
    editing.value = false
    await loadInstances()
    const found = instances.value.find((i: any) => i.id === id)
    if (found) {
      selected.value = found
      loadConfig(found)
    }
  } catch {}
}

async function doDelete() {
  if (!selected.value) return
  if (!confirm(t('instances.edit.confirm_delete', { name: displayName(selected.value) }))) return
  try {
    await window.InstancesManager.Delete(selected.value.id)
    selected.value = null
    view.value = 'list'
    await loadInstances()
  } catch {}
}

async function checkRunning() {
  if (!selected.value || !runningLaunchId.value) {
    isRunning.value = false
    return
  }
  try {
    const running = await window.NovaCoreManager.RunningInstances()
    isRunning.value = running.some((i: any) => {
      const id = i?.launchId ?? i?.id ?? i?.instanceId
      return id === runningLaunchId.value
    })
    if (!isRunning.value) {
      runningLaunchId.value = ''
      stopRunningPoll()
    }
  } catch {
    isRunning.value = false
  }
}

function startRunningPoll() {
  stopRunningPoll()
  runningPollTimer.value = setInterval(checkRunning, 5000)
}

function stopRunningPoll() {
  if (runningPollTimer.value) {
    clearInterval(runningPollTimer.value)
    runningPollTimer.value = null
  }
}

async function doKill() {
  if (!selected.value || !runningLaunchId.value) return
  try {
    await window.NovaCoreManager.KillInstance(runningLaunchId.value)
    runningLaunchId.value = ''
    isRunning.value = false
    stopRunningPoll()
    stopPlaytimeTracking()
  } catch {}
}

async function openInstanceFolder() {
  if (!selected.value) return
  await window.ElectronAPI.OpenPath(selected.value.dir)
}

async function doLaunch() {
  if (!selected.value) return
  launching.value = true
  try {
    const version = lastVersion(selected.value)
    if (!version || version === '?') return

    const instCfg = selected.value?.config
    const selectedVer = instCfg?.configInstance?.customFields?.selectedVersion
    const effectiveVersion = selectedVer || version

    const instancesDir = await window.InstancesManager.GetDir()
    const sharedDir = await window.InstancesManager.GetSharedDir()
    const token = await window.AuthManager.GetToken()
    const info = await window.NovaCoreManager.GetInfo()
    const cfg = await window.ConfigManager.Get()

    const jvmExtraArgs = (instCfg?.configInstance?.jvm?.extraArgs || []).filter(Boolean)
    const jvmPrependArgs = (instCfg?.configInstance?.jvm?.prependArgs || []).filter(Boolean)

    const result = await window.NovaCoreManager.Launch({
      version: effectiveVersion,
      instancePath: selected.value.dir,
      sharedPath: sharedDir,
      isInstance: true,
      javaPath: instCfg?.configInstance?.javaPath || undefined,
      hardwareAcceleration: instCfg?.configInstance?.hardwareAccel ?? undefined,
      gcPreset: instCfg?.configInstance?.gcPreset !== 'none' ? instCfg?.configInstance?.gcPreset : undefined,
      gpuPreference: instCfg?.configInstance?.gpuPreference || undefined,
      auth: {
        username: user.value?.username ?? 'Player',
        uuid: user.value?.uuid ?? '',
        accessToken: token ?? '',
        userType: 'mojang',
      },
      authlibInjector: {
        enabled: true,
        jarPath: info.authlibPath,
        serverUrl: 'https://api.stepnicka012.workers.dev',
      },
      jvm: {
        minMemoryMb: instCfg?.configInstance?.jvm?.minMemoryMb ?? instCfg?.configInstance?.minMemoryMb ?? 512,
        maxMemoryMb: instCfg?.configInstance?.jvm?.maxMemoryMb ?? instCfg?.configInstance?.maxMemoryMb ?? 2048,
        extraArgs: jvmExtraArgs,
        prependArgs: jvmPrependArgs,
      },
      window: {
        width: instCfg?.configInstance?.window?.width ?? 854,
        height: instCfg?.configInstance?.window?.height ?? 480,
        fullscreen: instCfg?.configInstance?.window?.fullscreen ?? false,
      },
      game: {
        gameDir: `${selected.value.dir}/game`,
        extraGameArgs: (instCfg?.configInstance?.extraGameArgs || []).filter(Boolean),
      },
      launcher: { name: 'StepLauncher', version: '1.0.0' },
    })

    if (!result.success) {
      console.error('Launch error:', result.error)
      return
    }

    runningLaunchId.value = result.launchId ?? ''
    isRunning.value = true
    startRunningPoll()
    startPlaytimeTracking()

    const cf = { ...instanceCF(selected.value), lastPlayedAt: new Date().toISOString() }
    window.InstancesManager.UpdateConfig(selected.value.id, { configInstance: { customFields: cf } }).catch(() => {})
    loadInstances()

    window.AuthManager.UpdateStats({
      last_version_playing: version,
      playing_time_increment: 0,
    }).catch(() => {})

    if (cfg.launcher.hideOnLaunch) {
      window.ElectronAPI.Hide()
      if (runningLaunchId.value) {
        let gameStarted = false
        const checkInterval = setInterval(async () => {
          try {
            const running = await window.NovaCoreManager.RunningInstances()
            const found = running.some((i: any) => {
              const id = i?.launchId ?? i?.id ?? i?.instanceId
              return id === runningLaunchId.value
            })
            if (found) {
              gameStarted = true
              return
            }
            if (gameStarted && !found) {
              clearInterval(checkInterval)
              stopPlaytimeTracking()
              updatePlaytime()
              window.ElectronAPI.Show()
              if (document.hidden)               window.ElectronAPI.ShowNotification({
                title: t('notifications.game.closed_title'),
                body: t('notifications.game.closed_inst_body', { name: displayName(selected.value) }),
              })
            }
          } catch {
            clearInterval(checkInterval)
          }
        }, 5000)
      }
    }
  } catch (err: any) {
    console.error('Launch error:', err)
  } finally {
    launching.value = false
  }
}

function updatePlaytime() {
  if (!selected.value || !runningLaunchId.value) return
  const elapsed = Date.now() - launchStartTime.value
  if (elapsed < 5000) return
  const current = selected.value.config?.instanceMetadata?.totalPlayTimeMs ?? 0
  const newTotal = current + elapsed
  window.InstancesManager.UpdateConfig(selected.value.id, {
    instanceMetadata: { totalPlayTimeMs: newTotal }
  }).catch(() => {})
  launchStartTime.value = Date.now()
}

function startPlaytimeTracking() {
  stopPlaytimeTracking()
  launchStartTime.value = Date.now()
  playtimeTimer = setInterval(() => {
    updatePlaytime()
  }, 30000)

  launchExitUnsub = window.NovaCoreManager.OnEvent((ev) => {
    if (ev.event === 'launch_exited') {
      const d = ev.data as any
      if (d.launchId === runningLaunchId.value) {
        stopPlaytimeTracking()
        updatePlaytime()
        setTimeout(() => loadInstances(), 1000)
      }
    }
  })
}

function stopPlaytimeTracking() {
  if (playtimeTimer) {
    clearInterval(playtimeTimer)
    playtimeTimer = null
  }
  if (launchExitUnsub) {
    launchExitUnsub()
    launchExitUnsub = null
  }
}

onMounted(async () => {
  try {
    const info = await window.NovaCoreManager.GetInfo()
    if (info?.javaPath) defaultJavaPath.value = info.javaPath
  } catch {}
  loadInstances()
  loadVersions()
  nextTick(() => {
    const el = document.querySelector('.instances-panel') as HTMLElement
    el?.focus()
  })
  if (selected.value) {
    try {
      const running = await window.NovaCoreManager.RunningInstances()
      const match = running.find((i: any) => i.instancePath === selected.value?.dir || i.dir === selected.value?.dir)
      if (match) {
        runningLaunchId.value = match?.launchId ?? match?.id ?? match?.instanceId ?? ''
        isRunning.value = true
        startRunningPoll()
      }
    } catch {}
  }
})

onUnmounted(() => {
  if (installUnsub) installUnsub()
  if (consoleUnsub) consoleUnsub()
  stopRunningPoll()
  stopPlaytimeTracking()
})
</script>

<style scoped lang="scss" src="../Styles/Panels/InstancesPanel.scss"></style>
