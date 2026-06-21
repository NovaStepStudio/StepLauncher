<template>
  <section class="mods-panel">
    <aside class="mods-sidebar">
      <button class="btn-icon" @click="$emit('close')">
        <img :src="'assets/svg/arrow-left.svg'" width="16" height="16" />
      </button>
      <h2 class="mods-title">{{ $t('mods.title') }}</h2>

      <div class="tabs">
        <button :class="['tab', { active: tab === 'browse' }]" @click="tab = 'browse'">{{ $t('mods.tabs.browse') }}</button>
        <button :class="['tab', { active: tab === 'library' }]" @click="tab = 'library'">{{ $t('mods.tabs.library') }}</button>
      </div>

      <template v-if="tab === 'browse'">
        <div class="search-wrap">
          <input v-model="query" class="search-input" :placeholder="$t('mods.browse.search_placeholder')" @input="onSearch" />
          <img class="search-icon" :src="'assets/svg/search.svg'" width="14" height="14" />
        </div>

        <div class="filter-section">
          <div class="filter-group">
            <span class="filter-label">{{ $t('mods.browse.filters.type') }}</span>
            <div class="filter-chips">
              <button v-for="typeItem in types" :key="typeItem.key" :class="['chip', { active: type === typeItem.key }]" @click="type = typeItem.key; doSearch()">{{ typeItem.label }}</button>
            </div>
          </div>

          <div class="filter-group" v-if="type === 'mod'">
            <span class="filter-label">{{ $t('mods.browse.filters.loader') }}</span>
            <div class="filter-chips">
              <button v-for="l in loaders" :key="l.key" :class="['chip', { active: loader === l.key }]" @click="loader = loader === l.key ? '' : l.key; doSearch()">{{ l.label }}</button>
            </div>
          </div>

          <div class="filter-row">
            <div class="filter-group">
              <span class="filter-label">{{ $t('mods.browse.filters.mc_version') }}</span>
              <select v-model="gameVersion" class="filter-select" @change="doSearch">
                <option value="">{{ $t('mods.browse.option_all') }}</option>
                <option v-for="v in gameVersions" :key="v" :value="v">{{ v }}</option>
              </select>
            </div>
            <div class="filter-group">
              <span class="filter-label">{{ $t('mods.browse.filters.sort') }}</span>
              <select v-model="sortIndex" class="filter-select" @change="doSearch">
                <option value="relevance">{{ $t('mods.browse.sort.relevance') }}</option>
                <option value="downloads">{{ $t('mods.browse.sort.downloads') }}</option>
                <option value="newest">{{ $t('mods.browse.sort.newest') }}</option>
                <option value="updated">{{ $t('mods.browse.sort.updated') }}</option>
              </select>
            </div>
          </div>

          <div class="filter-group">
            <span class="filter-label">{{ $t('mods.browse.filters.category') }}</span>
            <select v-model="category" class="filter-select" @change="doSearch">
              <option value="">{{ $t('mods.browse.option_all') }}</option>
              <option v-for="c in categories" :key="c.name" :value="c.name">{{ c.name }}</option>
            </select>
          </div>
        </div>
      </template>

      <template v-if="tab === 'library'">
        <button class="sidebar-create-btn" @click="showCreateLib = true">
          <img :src="'assets/svg/plus.svg'" width="12" height="12" />
          {{ $t('mods.library.new') }}
        </button>
        <div class="filter-group">
          <span class="filter-label">{{ $t('mods.browse.filters.sort') }}</span>
          <select v-model="libSort" class="filter-select">
            <option value="name">{{ $t('mods.library.sort_name') }}</option>
            <option value="entries">{{ $t('mods.library.sort_entries') }}</option>
            <option value="newest">{{ $t('mods.library.sort_newest') }}</option>
          </select>
        </div>
        <div v-if="activeLibrary" class="lib-sidebar-info">
          <span class="filter-label">{{ $t('mods.library.current') }}</span>
          <div class="lib-sidebar-current">
            <img :src="'assets/svg/layers.svg'" width="16" height="16" />
            <span>{{ activeLibrary.title }}</span>
          </div>
        </div>
        <div v-else class="lib-sidebar-hint">
          <img :src="'assets/svg/layers.svg'" width="18" height="18" style="opacity:0.2" />
          <span>{{ $t('mods.library.select_hint') }}</span>
        </div>
      </template>
    </aside>

    <main class="mods-main">
      <template v-if="detail">
        <button class="back-btn" @click="detail = null">
          <img :src="'assets/svg/arrow-left.svg'" width="14" height="14" />
          {{ $t('mods.library.back') }}
        </button>

        <div class="detail-header">
          <img v-if="detail.icon_url" :src="detail.icon_url" class="detail-icon" />
          <div class="detail-info">
            <h2 class="detail-title">{{ detail.title }}</h2>
            <span class="detail-author" v-if="detail.author">{{ detail.author }}</span>
            <div class="detail-stats">
              <span class="stat" v-if="detail.downloads != null">
                <img :src="'assets/svg/download.svg'" width="12" height="12" />
                {{ formatNumber(detail.downloads) }}
              </span>
              <span class="stat" v-if="detail.follows != null">
                <img :src="'assets/svg/user-group.svg'" width="12" height="12" />
                {{ formatNumber(detail.follows) }}
              </span>
              <span class="stat" v-if="detail.project_type">
                <img :src="'assets/svg/package.svg'" width="12" height="12" />
                {{ detail.project_type }}
              </span>
            </div>
            <div class="detail-acts">
              <button class="detail-act-btn" @click="showVersionPicker = true">
                <img :src="'assets/svg/download.svg'" width="12" height="12" />
                {{ $t('mods.detail.buttons.download') }}
              </button>
              <button class="detail-act-btn detail-act-btn--sec" @click="showAddToLib = true">
                <img :src="'assets/svg/plus.svg'" width="12" height="12" />
                {{ $t('mods.detail.buttons.add') }}
              </button>
            </div>
          </div>
        </div>

        <div class="detail-tabs">
          <button :class="['detail-tab', { active: detailTab === 'info' }]" @click="detailTab = 'info'">{{ $t('mods.detail.tabs.info') }}</button>
          <button :class="['detail-tab', { active: detailTab === 'versions' }]" @click="detailTab = 'versions'">{{ $t('mods.detail.tabs.versions') }}</button>
          <button v-if="detail.gallery?.length" :class="['detail-tab', { active: detailTab === 'gallery' }]" @click="detailTab = 'gallery'">{{ $t('mods.detail.tabs.gallery') }}</button>
        </div>

        <template v-if="detailTab === 'info'">
          <p class="detail-desc">{{ detail.description }}</p>

          <div class="detail-info-grid" v-if="detail.client_side || detail.categories?.length">
            <div class="info-tag" v-if="detail.client_side">
              <span class="tag-label">{{ $t('mods.detail.client_side') }}</span>
              <span class="tag-value">{{ detail.client_side }}</span>
            </div>
            <div class="info-tag" v-if="detail.categories?.length">
              <span class="tag-label">{{ $t('mods.detail.categories') }}</span>
              <span class="tag-value">{{ detail.categories.join(', ') }}</span>
            </div>
          </div>

          <div v-if="detail.body" class="detail-body"><div class="detail-body__markdown" v-html="detail.bodyHtml" /></div>
        </template>

        <template v-if="detailTab === 'versions'">
          <div v-if="loadingVersions" class="loading-spinner">
            <div class="spinner" />
            <span>{{ $t('mods.detail.versions.loading') }}</span>
          </div>
          <div v-else-if="!versions.length" class="state-msg">{{ $t('mods.detail.versions.empty') }}</div>
          <template v-else>
            <div class="version-loader-filter">
              <button :class="['vl-chip', { active: versionLoader === '' }]" @click="versionLoader = ''">{{ $t('mods.detail.versions.filter_all') }}</button>
              <button v-for="ld in versionLoaders" :key="ld" :class="['vl-chip', { active: versionLoader === ld }]" @click="versionLoader = ld">{{ ld }}</button>
            </div>
            <div class="version-search">
              <img class="search-icon" :src="'assets/svg/search.svg'" width="14" height="14" />
              <input class="version-search-input" type="text" v-model="versionSearch" :placeholder="$t('mods.detail.versions.search_placeholder')" />
            </div>
            <div class="version-list">
              <div v-for="v in filteredVersionsPaginated" :key="v.id" class="version-item">
                <div class="version-main">
                  <strong>{{ v.name || v.version_number }}</strong>
                  <div class="version-tags">
                    <span class="vtag" v-for="gv in v.game_versions?.slice(0, 3)" :key="gv">{{ gv }}</span>
                    <span class="vtag vtag--more" v-if="v.game_versions?.length > 3">+{{ v.game_versions.length - 3 }}</span>
                    <span class="vtag vtag--loader" v-for="ld in v.loaders" :key="ld">{{ ld }}</span>
                  </div>
                  <span class="version-date">{{ formatDate(v.date_published) }}</span>
                </div>
                <div class="version-files" v-if="v.files?.length">
                  <span class="file-size" v-if="v.files[0].size">{{ formatBytes(v.files[0].size) }}</span>
                  <button class="dl-btn" @click="downloadMod(v)" :disabled="downloading === v.id">
                    {{ downloading === v.id ? '...' : $t('mods.detail.versions.download') }}
                  </button>
                </div>
              </div>
            </div>
            <div v-if="filteredVersionPages > 1" class="version-pagination">
              <button class="page-btn" :disabled="versionPage <= 1" @click="versionPage--">
                <img :src="'assets/svg/chevron-left.svg'" width="10" height="10" />
              </button>
              <button v-for="p in visibleVersionPages" :key="p"
                :class="['page-btn', { active: versionPage === p, ellipsis: p === '...' }]"
                @click="p !== '...' && (versionPage = p as number)">{{ p }}</button>
              <button class="page-btn" :disabled="versionPage >= filteredVersionPages" @click="versionPage++">
                <img :src="'assets/svg/chevron-right.svg'" width="10" height="10" />
              </button>
            </div>
          </template>
        </template>

        <template v-if="detailTab === 'gallery' && detail.gallery?.length">
          <div class="gallery-grid">
            <div v-for="img in detail.gallery" :key="img.url" class="gallery-item" @click="openGallery(img.url)">
              <div class="gallery-item__img-wrap">
                <img :src="img.url" :alt="img.title || ''" loading="lazy" />
                <div class="gallery-item__static-light"></div>
              </div>
              <p v-if="img.description" class="gallery-item__desc">{{ img.description }}</p>
            </div>
          </div>
          <div v-if="versionPreviewImg" class="gallery-overlay" @click.self="closeGallery">
            <button class="gallery-overlay__close" @click="closeGallery">
              <img :src="'assets/svg/x.svg'" width="18" height="18" />
            </button>
            <div class="gallery-overlay__viewport" @wheel.prevent="onGalleryZoom" @mousedown.prevent="startGalleryDrag" @mousemove.prevent="onGalleryDrag" @mouseup.prevent="endGalleryDrag" @mouseleave.prevent="endGalleryDrag">
              <img :src="versionPreviewImg" class="gallery-overlay__img" :style="galleryImgStyle" draggable="false" />
            </div>
            <div class="gallery-overlay__controls">
              <button class="gallery-overlay__ctrl" @click="galleryZoomIn">+</button>
              <span class="gallery-overlay__zoom-lbl">{{ Math.round(galleryZoom * 100) }}%</span>
              <button class="gallery-overlay__ctrl" @click="galleryZoomOut">−</button>
              <button class="gallery-overlay__ctrl gallery-overlay__ctrl--reset" @click="resetGalleryZoom">↺</button>
            </div>
          </div>
        </template>
      </template>

      <template v-else-if="tab === 'browse'">
        <div v-if="loading" class="loading-spinner">
          <div class="spinner" />
          <span>{{ $t('mods.browse.loading') }}</span>
        </div>
        <div v-else-if="error" class="state-msg state-msg--error">{{ error }}</div>
        <div v-else-if="!results.length" class="state-msg">{{ $t('mods.browse.no_results') }}</div>

        <div v-else class="results-grid">
          <div v-for="p in results" :key="p.project_id" class="mod-card-grid" @click="openDetail(p)">
            <img v-if="p.icon_url" :src="p.icon_url" class="mod-card-grid__icon" loading="lazy" />
            <div class="mod-card-grid__body">
              <div class="mod-card-grid__title">
                <strong>{{ p.title }}</strong>
                <span class="mod-card-grid__type">{{ p.project_type }}</span>
              </div>
              <span class="mod-card-grid__author">{{ p.author || '—' }}</span>
              <p class="mod-card-grid__desc">{{ p.description }}</p>
              <div class="mod-card-grid__meta">
                <span class="mod-card-grid__stat">
                  <img :src="'assets/svg/download.svg'" width="12" height="12" />
                  {{ formatNumber(p.downloads) }}
                </span>
                <span class="mod-card-grid__stat">
                  <img :src="'assets/svg/user-group.svg'" width="12" height="12" />
                  {{ formatNumber(p.follows) }}
                </span>
                <div class="mod-card-grid__tags">
                  <span v-for="cat in (p.categories || []).slice(0, 2)" :key="cat" class="mod-card-grid__tag">{{ cat }}</span>
                </div>
              </div>
            </div>
            <div class="mod-card-grid__overlay">
              <button class="mod-card-grid__ol-btn" @click.stop="openDetail(p)">{{ $t('mods.browse.card_view') }}</button>
              <button class="mod-card-grid__ol-btn mod-card-grid__ol-btn--accent" @click.stop="showAddToLibFromCard(p)">+</button>
            </div>
          </div>
        </div>

        <div v-if="searchPages > 1" class="search-pagination">
          <button class="page-btn" :disabled="searchPage <= 1" @click="goToSearchPage(searchPage - 1)">
            <img :src="'assets/svg/chevron-left.svg'" width="10" height="10" />
          </button>
          <button v-for="p in visibleSearchPages" :key="p"
            :class="['page-btn', { active: searchPage === p, ellipsis: p === '...' }]"
            @click="p !== '...' && goToSearchPage(p as number)">{{ p }}</button>
          <button class="page-btn" :disabled="searchPage >= searchPages" @click="goToSearchPage(searchPage + 1)">
            <img :src="'assets/svg/chevron-right.svg'" width="10" height="10" />
          </button>
        </div>
      </template>

      <template v-else-if="tab === 'library'">
        <div v-if="loadingLib" class="loading-spinner">
          <div class="spinner" />
          <span>{{ $t('mods.library.loading') }}</span>
        </div>
        <div v-else-if="libError" class="state-msg state-msg--error">{{ libError }}</div>

        <template v-else-if="!activeLibrary">
          <div v-if="!libraries.length" class="state-msg">
            <img :src="'assets/svg/layers.svg'" width="32" height="32" style="opacity:0.12" />
            <span>{{ $t('mods.library.empty') }}</span>
          </div>
          <div v-else class="lib-grid">
            <div v-for="lib in sortedLibraries" :key="lib.id" class="lib-card" @click="openLibrary(lib)">
              <div class="lib-card__preview">
                <div v-if="lib._icons?.length" class="lib-card__grid">
                  <img v-for="(icon, i) in lib._icons.slice(0, 4)" :key="i" :src="icon" class="lib-card__thumb" />
                </div>
                <div v-else class="lib-card__grid lib-card__grid--empty">
                  <img :src="'assets/svg/layers.svg'" width="24" height="24" style="opacity:0.2" />
                </div>
              </div>
              <div class="lib-card__body">
                <h3 class="lib-card__title">{{ lib.title }}</h3>
                <p v-if="lib.description" class="lib-card__desc">{{ lib.description }}</p>
                <div class="lib-card__footer">
                  <span>{{ $t('mods.library.entries_count', { count: lib._entries?.length || 0 }) }}</span>
                </div>
              </div>
            </div>
          </div>
        </template>

        <template v-else>
          <button class="back-btn" @click="activeLibrary = null">
            <img :src="'assets/svg/arrow-left.svg'" width="14" height="14" />
            {{ $t('mods.library.back_to_libraries') }}
          </button>
          <h3 class="lib-entries-title">{{ activeLibrary.title }}</h3>
          <div v-if="!activeEntries.length" class="state-msg">{{ $t('mods.library.empty_entries') }}</div>
          <div v-else class="results-list">
            <div v-for="e in activeEntries" :key="e.id || e.project_id" class="mod-card" @click="openLibraryDetail(e)">
              <img v-if="e.icon" :src="e.icon" class="mod-card__icon" loading="lazy" />
              <div v-else class="mod-card__icon mod-card__icon--empty" />
              <div class="mod-info">
                <strong class="mod-title">{{ e.title }}</strong>
                <span class="mod-desc">{{ e.description }}</span>
                <div class="mod-meta">
                  <span class="meta-type">{{ e.entry_type || e.project_type }}</span>
                  <span v-if="e.author_name" class="meta-author">{{ e.author_name }}</span>
                </div>
              </div>
              <button class="remove-btn" @click.stop="removeEntry(e)" :title="$t('mods.library.remove')">
                <img :src="'assets/svg/x.svg'" width="12" height="12" />
              </button>
            </div>
          </div>
        </template>
      </template>
    </main>

    <!-- ── Version Picker Modal ── -->
    <div v-if="showVersionPicker" class="modal-overlay" @click.self="showVersionPicker = false">
      <div class="version-picker-modal">
        <div class="version-picker-modal__head">
          <h3>{{ $t('mods.modals.select_version') }}</h3>
          <button class="modal-close-btn" @click="showVersionPicker = false">
            <img :src="'assets/svg/x.svg'" width="16" height="16" />
          </button>
        </div>
        <div class="version-loader-filter">
          <button :class="['vl-chip', { active: pickerVersionLoader === '' }]" @click="pickerVersionLoader = ''">{{ $t('mods.detail.versions.filter_all') }}</button>
          <button v-for="ld in pickerVersionLoaders" :key="ld" :class="['vl-chip', { active: pickerVersionLoader === ld }]" @click="pickerVersionLoader = ld">{{ ld }}</button>
        </div>
        <div class="picker-search">
          <img class="picker-search-icon" :src="'assets/svg/search.svg'" width="12" height="12" />
          <input v-model="pickerSearch" class="picker-search-input" :placeholder="$t('mods.detail.versions.search_placeholder')" />
        </div>
        <div v-if="loadingVersions" class="modal-loading"><div class="spinner" /></div>
        <div v-else class="version-picker-modal__list">
          <div v-for="v in pickerPaginatedVersions" :key="v.id" class="version-item">
            <div class="version-main">
              <strong>{{ v.name || v.version_number }}</strong>
              <div class="version-tags">
                <span class="vtag" v-for="gv in v.game_versions?.slice(0, 3)" :key="gv">{{ gv }}</span>
                <span class="vtag vtag--more" v-if="v.game_versions?.length > 3">+{{ v.game_versions.length - 3 }}</span>
                <span class="vtag vtag--loader" v-for="ld in v.loaders" :key="ld">{{ ld }}</span>
              </div>
              <span class="version-date">{{ formatDate(v.date_published) }}</span>
            </div>
            <div class="version-files" v-if="v.files?.length">
              <span class="file-size" v-if="v.files[0].size">{{ formatBytes(v.files[0].size) }}</span>
              <button class="dl-btn" @click="downloadMod(v)" :disabled="downloading === v.id">
                    {{ downloading === v.id ? '...' : $t('mods.detail.versions.download') }}
              </button>
            </div>
          </div>
        </div>
        <div v-if="pickerPages > 1" class="picker-pagination">
          <button class="page-btn" :disabled="pickerPage <= 1" @click="pickerPage--">&lsaquo;</button>
          <button v-for="p in pickerVisiblePages" :key="p"
            :class="['page-btn', { active: pickerPage === p }]"
            @click="p !== '...' && (pickerPage = p as number)">{{ p }}</button>
          <button class="page-btn" :disabled="pickerPage >= pickerPages" @click="pickerPage++">&rsaquo;</button>
        </div>
      </div>
    </div>

    <!-- ── Add to Library Modal ── -->
    <div v-if="showAddToLib" class="modal-overlay" @click.self="showAddToLib = false">
      <div class="lib-picker-modal">
        <div class="version-picker-modal__head">
          <h3>{{ $t('mods.modals.add_to_library') }}</h3>
          <button class="modal-close-btn" @click="showAddToLib = false">
            <img :src="'assets/svg/x.svg'" width="16" height="16" />
          </button>
        </div>
        <div v-if="libPicking" class="modal-loading"><div class="spinner" /></div>
        <div v-else-if="!libPickerLibs.length" class="modal-state-msg">{{ $t('mods.modals.no_libraries') }}</div>
        <div v-else class="lib-picker-modal__list">
          <div v-for="lib in libPickerLibs" :key="lib.id" class="lib-picker-item" @click="doAddEntry(lib)">
            <div class="lib-picker-item__preview">
              <div v-if="lib._icons?.length" class="lib-picker-item__grid">
                <img v-for="(icon, i) in lib._icons.slice(0, 4)" :key="i" :src="icon" class="lib-picker-item__thumb" />
              </div>
              <div v-else class="lib-picker-item__grid lib-picker-item__grid--empty">
                <img :src="'assets/svg/layers.svg'" width="16" height="16" />
              </div>
            </div>
            <div class="lib-picker-item__info">
              <strong>{{ lib.title }}</strong>
              <span>{{ $t('mods.library.entries_count', { count: lib._entries?.length || 0 }) }}</span>
            </div>
            <span v-if="addingToLibId === lib.id" class="lib-picker-item__add">⋯</span>
            <span v-else class="lib-picker-item__add">+</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ── Create Library Modal ── -->
    <div v-if="showCreateLib" class="modal-overlay" @click.self="showCreateLib = false">
      <div class="create-lib-modal">
        <div class="version-picker-modal__head">
          <h3>{{ $t('mods.modals.new_library') }}</h3>
          <button class="modal-close-btn" @click="showCreateLib = false">
            <img :src="'assets/svg/x.svg'" width="16" height="16" />
          </button>
        </div>
        <div class="create-lib-modal__body">
          <input v-model="createLibForm.title" :placeholder="$t('mods.modals.title_placeholder')" class="create-lib-modal__input" maxlength="100" />
          <textarea v-model="createLibForm.desc" :placeholder="$t('mods.modals.desc_placeholder')" class="create-lib-modal__input create-lib-modal__input--ta" maxlength="1000" rows="2"></textarea>
          <label class="create-lib-modal__toggle">
            <input type="checkbox" v-model="createLibForm.isPublic" />
            {{ $t('mods.modals.is_public') }}
          </label>
          <div class="create-lib-modal__acts">
            <button class="modal-cancel-btn" @click="showCreateLib = false">{{ $t('mods.modals.cancel') }}</button>
            <button class="modal-primary-btn" @click="doCreateLib" :disabled="!createLibForm.title.trim() || creatingLib">{{ $t('mods.modals.create') }}</button>
          </div>
        </div>
      </div>
    </div>

    <!-- ── World Picker Modal ── -->
    <div v-if="showWorldPicker" class="modal-overlay" @click.self="showWorldPicker = false">
      <div class="modal-card world-picker-modal">
        <div class="modal-header">
          <span>{{ $t('mods.modals.world_picker') }}</span>
          <button class="modal-close-btn" @click="showWorldPicker = false">&times;</button>
        </div>
        <div class="world-picker-body">
          <p class="world-picker-desc">{{ $t('mods.modals.world_picker_desc') }}</p>
          <div class="world-list">
            <button v-for="w in worlds" :key="w" class="world-item" @click="downloadDatapackToWorld(w)">
              <img :src="'assets/svg/package.svg'" width="16" height="16" />
              {{ w }}
            </button>
          </div>
          <button class="world-item world-item--all" @click="downloadDatapackGlobal">
            <img :src="'assets/svg/globe.svg'" width="16" height="16" />
            {{ $t('mods.modals.install_all_worlds') }}
          </button>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { t } from '../i18n'
import { useNotifications } from '../Composables/useNotifications'
defineEmits<{ (e: 'close'): void }>()
const tab = ref('browse')

const types = [
  { key: 'mod', label: t('mods.browse.types.mods') },
  { key: 'shader', label: t('mods.browse.types.shaders') },
  { key: 'resourcepack', label: t('mods.browse.types.resourcepacks') },
  { key: 'modpack', label: t('mods.browse.types.modpacks') },
  { key: 'datapack', label: t('mods.browse.types.datapacks') },
]
const loaders = [
  { key: 'fabric', label: t('mods.browse.loaders.fabric') },
  { key: 'forge', label: t('mods.browse.loaders.forge') },
  { key: 'neoforge', label: t('mods.browse.loaders.neoforge') },
  { key: 'quilt', label: t('mods.browse.loaders.quilt') },
]

const query = ref('')
const type = ref('mod')
const loader = ref('')
const gameVersion = ref('')
const sortIndex = ref('relevance')
const category = ref('')

const gameVersions = ref<string[]>([])
const categories = ref<{ name: string; project_type: string }[]>([])

const results = ref<any[]>([])
const loading = ref(false)
const error = ref('')
const searchPage = ref(1)
const totalResults = ref(0)
const perPage = 24

let searchTimer: ReturnType<typeof setTimeout> | null = null

const detail = ref<any>(null)
const detailTab = ref('info')
const versions = ref<any[]>([])
const versionPage = ref(1)
const versionsPerPage = 15
const versionLoader = ref('')
const versionSearch = ref('')
const versionPreviewImg = ref<string | null>(null)
const galleryZoom = ref(1)
const galleryPanX = ref(0)
const galleryPanY = ref(0)
const galleryDragging = ref(false)
const galleryDragStart = ref({ x: 0, y: 0 })

const showVersionPicker = ref(false)
const showAddToLib = ref(false)
const showCreateLib = ref(false)
const pickerVersionLoader = ref('')
const pickerSearch = ref('')
const pickerPage = ref(1)
const pickerPerPage = 15
const libPicking = ref(false)
const addingToLibId = ref<number | null>(null)
const libPickerLibs = ref<any[]>([])
const createLibForm = ref({ title: '', desc: '', isPublic: true })
const creatingLib = ref(false)

const { success: notifySuccess, error: notifyError, info: notifyInfo } = useNotifications()

const showWorldPicker = ref(false)
const worlds = ref<string[]>([])
const pendingWorldDownload = ref<any>(null)

const searchPages = computed(() => Math.ceil(totalResults.value / perPage))

const visibleSearchPages = computed(() => {
  const total = searchPages.value
  const current = searchPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: (number | string)[] = []
  if (current <= 4) {
    for (let i = 1; i <= Math.min(5, total); i++) pages.push(i)
    if (total > 5) { pages.push('...'); pages.push(total) }
  } else if (current >= total - 3) {
    pages.push(1); pages.push('...')
    for (let i = Math.max(total - 4, 1); i <= total; i++) pages.push(i)
  } else {
    pages.push(1); pages.push('...')
    for (let i = current - 1; i <= current + 1; i++) pages.push(i)
    pages.push('...'); pages.push(total)
  }
  return pages
})

const galleryImgStyle = computed(() => ({
  transform: `translate(${galleryPanX.value}px, ${galleryPanY.value}px) scale(${galleryZoom.value})`,
  cursor: galleryZoom.value > 1 ? 'grab' : 'default',
}))

const loadingVersions = ref(false)
const downloading = ref('')

const versionLoaders = computed(() => {
  const set = new Set<string>()
  for (const v of versions.value) {
    for (const ld of v.loaders || []) set.add(ld)
  }
  return [...set].sort()
})

const filteredVersions = computed(() => {
  let list = versions.value
  if (versionLoader.value) {
    list = list.filter((v: any) => (v.loaders || []).includes(versionLoader.value))
  }
  if (versionSearch.value) {
    const q = versionSearch.value.toLowerCase()
    list = list.filter((v: any) =>
      (v.name || v.version_number || '').toLowerCase().includes(q) ||
      (v.game_versions || []).some((gv: string) => gv.toLowerCase().includes(q))
    )
  }
  return list
})

const filteredVersionsPaginated = computed(() => {
  const start = (versionPage.value - 1) * versionsPerPage
  return filteredVersions.value.slice(start, start + versionsPerPage)
})

const filteredVersionPages = computed(() => Math.ceil(filteredVersions.value.length / versionsPerPage))

const pickerVersionLoaders = computed(() => {
  const set = new Set<string>()
  for (const v of versions.value) {
    for (const ld of v.loaders || []) set.add(ld)
  }
  return [...set].sort()
})

const pickerFilteredVersions = computed(() => {
  let list = versions.value
  if (pickerVersionLoader.value) list = list.filter((v: any) => (v.loaders || []).includes(pickerVersionLoader.value))
  if (pickerSearch.value.trim()) {
    const q = pickerSearch.value.toLowerCase().trim()
    list = list.filter((v: any) =>
      (v.name || v.version_number || '').toLowerCase().includes(q) ||
      (v.game_versions || []).some((gv: string) => gv.toLowerCase().includes(q))
    )
  }
  return list
})

const pickerPages = computed(() => Math.ceil(pickerFilteredVersions.value.length / pickerPerPage))

const pickerPaginatedVersions = computed(() => {
  const start = (pickerPage.value - 1) * pickerPerPage
  return pickerFilteredVersions.value.slice(start, start + pickerPerPage)
})

const pickerVisiblePages = computed(() => {
  const total = pickerPages.value
  const current = pickerPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: (number | string)[] = []
  if (current <= 4) {
    for (let i = 1; i <= Math.min(5, total); i++) pages.push(i)
    if (total > 5) { pages.push('...'); pages.push(total) }
  } else if (current >= total - 3) {
    pages.push(1); pages.push('...')
    for (let i = Math.max(total - 4, 1); i <= total; i++) pages.push(i)
  } else {
    pages.push(1); pages.push('...')
    for (let i = current - 1; i <= current + 1; i++) pages.push(i)
    pages.push('...'); pages.push(total)
  }
  return pages
})

const visibleVersionPages = computed(() => {
  const total = filteredVersionPages.value
  const current = versionPage.value
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const pages: (number | string)[] = []
  if (current <= 4) {
    for (let i = 1; i <= Math.min(5, total); i++) pages.push(i)
    if (total > 5) { pages.push('...'); pages.push(total) }
  } else if (current >= total - 3) {
    pages.push(1); pages.push('...')
    for (let i = Math.max(total - 4, 1); i <= total; i++) pages.push(i)
  } else {
    pages.push(1); pages.push('...')
    for (let i = current - 1; i <= current + 1; i++) pages.push(i)
    pages.push('...'); pages.push(total)
  }
  return pages
})

const libraries = ref<any[]>([])
const allEntries = ref<any[]>([])
const loadingLib = ref(false)
const libError = ref('')
const activeLibrary = ref<any>(null)
const libSort = ref('name')

const sortedLibraries = computed(() => {
  const items = [...libraries.value]
  if (libSort.value === 'name') {
    items.sort((a, b) => a.title.localeCompare(b.title))
  } else if (libSort.value === 'entries') {
    items.sort((a, b) => (b._entries?.length || 0) - (a._entries?.length || 0))
  } else if (libSort.value === 'newest') {
    items.sort((a, b) => ((b as any).created_at || '').localeCompare((a as any).created_at || ''))
  }
  return items
})

const activeEntries = computed(() => {
  if (!activeLibrary.value) return []
  const lib = libraries.value.find(l => l.id === activeLibrary.value.id)
  return lib?._entries || []
})

function formatNumber(n: number): string {
  if (!n) return '0'
  if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
  if (n >= 1e3) return (n / 1e3).toFixed(1) + 'k'
  return String(n)
}

function formatBytes(bytes: number): string {
  if (!bytes) return ''
  if (bytes >= 1e9) return (bytes / 1e9).toFixed(1) + ' GB'
  if (bytes >= 1e6) return (bytes / 1e6).toFixed(0) + ' MB'
  if (bytes >= 1e3) return (bytes / 1e3).toFixed(0) + ' KB'
  return bytes + ' B'
}

function formatDate(d: string): string {
  if (!d) return ''
  try {
    return new Date(d).toLocaleDateString('es-ES', { year: 'numeric', month: 'short', day: 'numeric' })
  } catch { return '' }
}

function onSearch() {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => doSearch(), 400)
}

function buildFacets(): string[][] {
  const f: string[][] = [[`project_type:${type.value}`]]
  if (loader.value) f.push([`categories:${loader.value}`])
  if (gameVersion.value) f.push([`versions:${gameVersion.value}`])
  if (category.value) f.push([`categories:${category.value}`])
  return f
}

async function doSearch() {
  loading.value = true; error.value = ''; searchPage.value = 1
  try {
    await runSearch(1)
  } catch (e: any) {
    error.value = e?.message || t('mods.browse.error')
    results.value = []
  }
  loading.value = false
}

async function runSearch(page: number) {
  const params = new URLSearchParams()
  if (query.value.trim()) params.set('query', query.value.trim())
  params.set('limit', String(perPage))
  params.set('offset', String((page - 1) * perPage))
  params.set('index', sortIndex.value)
  params.set('facets', JSON.stringify(buildFacets()))
  const res = await fetch(`https://api.modrinth.com/v2/search?${params}`)
  if (!res.ok) throw new Error(`Error ${res.status}`)
  const data = await res.json()
  results.value = data.hits || []
  totalResults.value = data.total_hits || 0
}

async function goToSearchPage(page: number) {
  if (page < 1 || page > searchPages.value || loading.value) return
  searchPage.value = page
  loading.value = true
  try {
    await runSearch(page)
  } catch {
    error.value = t('mods.browse.error')
  }
  loading.value = false
  window.scrollTo(0, 0)
}

async function openDetail(p: any) {
  detail.value = null; versions.value = []; detailTab.value = 'info'; versionPage.value = 1
  try {
    const res = await fetch(`https://api.modrinth.com/v2/project/${p.project_id || p.id}`)
    if (res.ok) {
      const data = await res.json()
      if (data.body) {
        const { marked } = await import('marked')
        data.bodyHtml = await marked.parse(data.body)
      }
      detail.value = data
      fetchVersions(data.id || data.project_id)
    }
  } catch {}
}

async function fetchVersions(projectId: string) {
  loadingVersions.value = true; versionPage.value = 1
  try {
    const res = await fetch(`https://api.modrinth.com/v2/project/${projectId}/version`)
    if (res.ok) {
      const data = await res.json()
      versions.value = data || []
    }
  } catch {}
  loadingVersions.value = false
}

async function downloadMod(v: any) {
  const primary = v.files?.[0]
  if (!primary) return
  downloading.value = v.id
  try {
    if (type.value === 'modpack') {
      const res = await window.ModsManager.InstallModpack(primary.url, primary.filename)
      if (res.success) {
        notifySuccess(t('mods.notifications.modpack_installed'))
      } else {
        notifyError(res.error || t('mods.notifications.modpack_error'))
      }
    } else if (type.value === 'datapack') {
      const worldsRes = await window.ModsManager.GetWorlds()
      if (worldsRes.success && worldsRes.worlds && worldsRes.worlds.length > 0) {
        worlds.value = worldsRes.worlds
        pendingWorldDownload.value = v
        showWorldPicker.value = true
      } else {
        const res = await window.ModsManager.DownloadToGameDir('datapack', primary.url, primary.filename)
        if (res.success) {
          notifySuccess(t('mods.notifications.datapack_global'))
        } else {
          notifyError(res.error || t('mods.notifications.download_error'))
        }
      }
    } else {
      const res = await window.ModsManager.DownloadToGameDir(type.value, primary.url, primary.filename)
      if (res.success) {
        notifySuccess(t('mods.notifications.mod_downloaded', { type: type.value }))
      } else {
        notifyError(res.error || t('mods.notifications.download_error'))
      }
    }
  } catch {}
  downloading.value = ''
}

async function downloadDatapackToWorld(world: string) {
  const v = pendingWorldDownload.value
  if (!v) return
  const primary = v.files?.[0]
  if (!primary) return
  showWorldPicker.value = false
  downloading.value = v.id
  try {
    const res = await window.ModsManager.DownloadToGameDir('datapack', primary.url, primary.filename, world)
    if (res.success) {
      notifySuccess(t('mods.notifications.datapack_world', { world }))
    } else {
      notifyError(res.error || t('mods.notifications.download_error'))
    }
  } catch {}
  downloading.value = ''
}

async function downloadDatapackGlobal() {
  const v = pendingWorldDownload.value
  if (!v) return
  const primary = v.files?.[0]
  if (!primary) return
  showWorldPicker.value = false
  downloading.value = v.id
  try {
    const res = await window.ModsManager.DownloadToGameDir('datapack', primary.url, primary.filename)
    if (res.success) {
      notifySuccess(t('mods.notifications.datapack_all_worlds'))
    } else {
      notifyError(res.error || t('mods.notifications.download_error'))
    }
  } catch {}
  downloading.value = ''
}

async function openLibraryDetail(e: any) {
  if (e.provider === 'modrinth' && e.project_id) {
    try {
      const res = await fetch(`https://api.modrinth.com/v2/project/${e.project_id}`)
      if (res.ok) {
        const data = await res.json()
        if (data.body) {
          const { marked } = await import('marked')
          data.bodyHtml = await marked.parse(data.body)
        }
        detail.value = data
        fetchVersions(data.id || data.project_id)
      }
    } catch {}
  }
}

async function removeEntry(e: any) {
  if (!e.id && !e.entry_id) return
  try {
    await window.AuthManager.RemoveEntry(e.library_id, e.id || e.entry_id)
    allEntries.value = allEntries.value.filter((x: any) => x.id !== e.id)
  } catch {}
}

async function loadLibrary() {
  loadingLib.value = true; libError.value = ''; activeLibrary.value = null
  try {
    const libs = await window.AuthManager.GetLibraries()
    libraries.value = Array.isArray(libs) ? libs : []

    const all: any[] = []
    for (const lib of libraries.value) {
      try {
        const entries = await window.AuthManager.GetEntries(lib.id)
        if (Array.isArray(entries)) {
          (lib as any)._entries = entries
          const icons = entries.filter((e: any) => e.icon).slice(0, 4).map((e: any) => e.icon)
          ;(lib as any)._icons = icons
          all.push(...entries.map((e: any) => ({ ...e, library_id: lib.id })))
        } else {
          (lib as any)._entries = []
          ;(lib as any)._icons = []
        }
      } catch {
        (lib as any)._entries = []
        ;(lib as any)._icons = []
      }
    }
    allEntries.value = all
  } catch (e: any) {
    libError.value = t('mods.library.error')
    libraries.value = []
    allEntries.value = []
  }
  loadingLib.value = false
}

function openLibrary(lib: any) {
  activeLibrary.value = lib
}

async function addToLibrary(p: any) {
  showAddToLib.value = true
  await loadLibPicker()
}

async function showAddToLibFromCard(p: any) {
  detail.value = p
  showAddToLib.value = true
  await loadLibPicker()
}

async function loadLibPicker() {
  libPicking.value = true
  try {
    const libs = await window.AuthManager.GetLibraries()
    libPickerLibs.value = Array.isArray(libs) ? libs : []
    for (const lib of libPickerLibs.value) {
      try {
        const entries = await window.AuthManager.GetEntries(lib.id)
        if (Array.isArray(entries)) {
          (lib as any)._entries = entries
          const icons = entries.filter((e: any) => e.icon).slice(0, 4).map((e: any) => e.icon)
          ;(lib as any)._icons = icons
        } else {
          (lib as any)._entries = []
          ;(lib as any)._icons = []
        }
      } catch {
        (lib as any)._entries = []
        ;(lib as any)._icons = []
      }
    }
  } catch {
    libPickerLibs.value = []
  }
  libPicking.value = false
}

async function doAddEntry(lib: any) {
  if (!detail.value || addingToLibId.value) return
  addingToLibId.value = lib.id
  const p = detail.value
  try {
    const entry: Record<string, any> = {
      project_id: p.id || p.project_id,
      provider: 'modrinth',
      entry_type: p.project_type || 'mod',
      slug: p.slug,
      title: p.title,
      description: p.description,
      icon: p.icon_url || null,
      color: p.color ? '#' + p.color.toString(16).toUpperCase().padStart(6, '0') : null,
      author_name: p.author || null,
      project_url: p.project_type && p.slug ? `https://modrinth.com/${p.project_type}/${p.slug}` : null,
      categories: p.categories || [],
      tags: [p.project_type || 'mod'],
    }
    await window.AuthManager.AddEntry(lib.id, entry)
    showAddToLib.value = false
  } catch {}
  addingToLibId.value = null
}

async function doCreateLib() {
  if (!createLibForm.value.title.trim() || creatingLib.value) return
  creatingLib.value = true
  try {
    await window.AuthManager.CreateLibrary(
      createLibForm.value.title.trim(),
      createLibForm.value.desc || null,
      createLibForm.value.isPublic
    )
    showCreateLib.value = false
    createLibForm.value = { title: '', desc: '', isPublic: true }
    loadLibrary()
  } catch {}
  creatingLib.value = false
}

function openGallery(url: string) {
  versionPreviewImg.value = url
  resetGalleryZoom()
}

function closeGallery() {
  versionPreviewImg.value = null
  resetGalleryZoom()
}

function resetGalleryZoom() {
  galleryZoom.value = 1
  galleryPanX.value = 0
  galleryPanY.value = 0
}

const ZOOM_STEP = 0.25
const ZOOM_MIN = 0.5
const ZOOM_MAX = 5

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

async function loadTags() {
  try {
    const [catRes, verRes] = await Promise.all([
      fetch('https://api.modrinth.com/v2/tag/category'),
      fetch('https://api.modrinth.com/v2/tag/game_version'),
    ])
    if (catRes.ok) {
      const data = await catRes.json()
      categories.value = (data || []).filter((c: any) => c.project_type === type.value)
    }
    if (verRes.ok) {
      const data = await verRes.json()
      gameVersions.value = (data || [])
        .filter((v: any) => v.version_type === 'release')
        .map((v: any) => v.version)
    }
  } catch {}
}

onMounted(() => {
  doSearch()
  loadTags()
})

watch(type, () => {
  loadTags()
  doSearch()
})

watch(tab, (t) => {
  detail.value = null
  if (t === 'library') loadLibrary()
})
watch([pickerVersionLoader, pickerSearch], () => { pickerPage.value = 1 })
</script>

<style scoped lang="scss" src="../Styles/Panels/ModsPanel.scss"></style>
