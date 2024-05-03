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
        content: "/contents/{id}/",
        contentScreens: base + 'contents/{mcID}/screens',
        screens: base + 'screens/',
        contentAll: base + 'content/',
        search: base + "search",

        // Task Related APIs
        requestScreens: "/editing_queue/{id}/screens/{count}/{startTimeSeconds}",
        encodeVideoContent: "/editing_queue/{id}/encoding",
        createPreviewFromScreens: "/editing_queue/{id}/webp",
        createTagContentTask: "/editing_queue/{id}/tagging",
        tags: base + "tags/",
    },
    tasks: {
        get: "/task_requests/{id}",
        update: "/task_requests/{id}",
        list: "/task_requests/",
    }
};
