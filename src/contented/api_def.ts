let base = window.location.origin + '/';
export let ApiDef = {
    base: base,
    contented: {
        preview: base + 'content/',
        fulldir: base + 'content/{dir}',
        download: base + 'download/{id}/{filename}'
    }
};
