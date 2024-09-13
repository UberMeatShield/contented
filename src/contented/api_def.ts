let base = window.location.origin + '/api/';

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
    content: '/api/contents/{id}/',
    contentScreens: base + 'contents/{mcID}/screens',
    screens: base + 'screens/',
    contentAll: base + 'content/',
    searchContents: '/api/search/contents',
    searchContainers: '/api/search/containers',
    tags: base + 'tags/',

    // Task Related APIs
    requestScreens: '/api/editing_queue/{id}/screens/{count}/{startTimeSeconds}',
    encodeVideoContent: '/api/editing_queue/{id}/encoding',
    createPreviewFromScreens: '/api/editing_queue/{id}/webp',
    createTagContentTask: '/api/editing_queue/{id}/tagging',
    contentDuplicatesTask: '/api/editing_queue/{contentId}/duplicates',

    // These will attempt to queue up tasks for ALL content in the container (but not in sub-containers)
    containerVideoEncodingTask: '/api/editing_container_queue/{containerId}/encoding',
    containerDuplicatesTask: '/api/editing_container_queue/{containerId}/duplicates',
    containerTaggingTask: '/api/editing_container_queue/{containerId}/tagging',
    containerPreviewsTask: '/api/editing_container_queue/{containerId}/screens/{count}/{startTimeSeconds}',
  },
  tasks: {
    get: base + 'task_requests/{id}',
    list: base + 'task_requests/',
    update: '/api/task_requests/{id}',
  },
};
