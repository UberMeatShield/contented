let base = window.location.origin + '/';

// Pagination is per_page and page to work with the standard Soda and Resource interfaces.
export let ApiDef = {
    base: base,
    contented: {
        preview: base + 'preview/',
        full: base + 'full/',
        download: base + 'download/',
        fulldir: base + 'containers/{dirId}/media',
        contents: base + 'containers/'
    }
};
