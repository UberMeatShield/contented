let base = window.location.origin + '/';

// Pagination is per_page and page to work with the standard Soda and Resource interfaces.
export let ApiDef = {
  base: base,
  contented: {
    splash: base + 'splash/',
    view: base + 'view/',
    download: base + 'download/{mcID}',
    preview: base + 'preview/',
    containers: base + 'containers/',
    containerContent: base + 'containers/{cId}/contents',
    content: '/contents/{id}/',
    contentScreens: base + 'contents/{mcID}/screens',
    screens: base + 'screens/',
    contentAll: base + 'content/',
    searchContents: base + 'api/search/contents',
    searchContainers: base + 'api/search/containers',
    tags: base + 'tags/',

    // Task Related APIs
    requestScreens: '/editing_queue/{id}/screens/{count}/{startTimeSeconds}',
    encodeVideoContent: '/editing_queue/{id}/encoding',
    createPreviewFromScreens: '/editing_queue/{id}/webp',
    createTagContentTask: '/editing_queue/{id}/tagging',
    contentDuplicatesTask: '/editing_queue/{contentId}/duplicates',

    // These will attempt to queue up tasks for ALL content in the container (but not in sub-containers)
    containerVideoEncodingTask: '/editing_container_queue/{containerId}/encoding',
    containerDuplicatesTask: '/editing_container_queue/{containerId}/duplicates',
    containerTaggingTask: '/editing_container_queue/{containerId}/tagging',
    containerPreviewsTask: '/editing_container_queue/{containerId}/screens/{count}/{startTimeSeconds}',
  },
  tasks: {
    get: '/task_requests/{id}',
    update: '/task_requests/{id}',
    list: '/task_requests/',
  },
};
