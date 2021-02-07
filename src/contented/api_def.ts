let base = window.location.origin + '/';

// Pagination is per_page and page to work with the standard Soda and Resource interfaces.
export let ApiDef = {
    base: base,
    contented: {
        view: base + 'view/',
        download: base + 'download/',
        preview: base + 'preview/',
        containers: base + 'containers/',
        media: base + 'containers/{dirId}/media',
    }
};
