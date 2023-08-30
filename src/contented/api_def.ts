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
        content: base + 'containers/{cId}/content',
        contentScreens: base + 'content/{mcID}/screens',
        screens: base + 'screens/',
        contentAll: base + 'content/',
        search: base + "search",
    }
};
