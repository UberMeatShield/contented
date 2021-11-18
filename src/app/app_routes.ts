import {NgModule} from '@angular/core';

import {ContentedCmp} from './../contented/contented_cmp';
import {SearchCmp} from './../contented/search_cmp';
import {APP_BASE_HREF} from '@angular/common';
import {RouterModule, Routes} from '@angular/router';

// Hmm, should have made this route have a saner extension
const appRoutes = [
    {path: '', redirectTo: '/ui/browse/0/0', pathMatch: 'full'},
    {path: 'ui/browse/:idx/:rowIdx', component: ContentedCmp},
    {path: 'ui/search', component: SearchCmp}
];
@NgModule({
    imports: [RouterModule.forRoot(appRoutes, { relativeLinkResolution: 'legacy' })],
    providers: [{provide: APP_BASE_HREF, useValue: ''}],
    exports: [RouterModule]
})
export class AppRoutes {}
