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
        containerContent: base + 'containers/{cId}/content',
        content: "/content/{id}/",
        contentScreens: base + 'content/{mcID}/screens',
        screens: base + 'screens/',
        requestScreens: "/editing_queue/{id}/screens/{count}/{startTimeSeconds}",
        contentAll: base + 'content/',
        search: base + "search",

        tags: base + "tags/",
    }
};
