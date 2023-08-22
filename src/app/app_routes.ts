import {NgModule} from '@angular/core';

import {ContentedCmp} from './../contented/contented_cmp';
import {SearchCmp} from './../contented/search_cmp';
import {VideoViewCmp} from './../contented/video_view_cmp';
import {ContentViewCmp} from './../contented/content_view_cmp';
import {APP_BASE_HREF} from '@angular/common';
import {RouterModule, Routes} from '@angular/router';

// Hmm, should have made this route have a saner extension
const appRoutes: Routes = [
    {path: '', redirectTo: '/ui/browse/0/0', pathMatch: 'full'},
    {path: 'ui/browse/:idx/:rowIdx', component: ContentedCmp},
    {path: 'ui/content/:id', component: ContentViewCmp},
    {path: 'ui/search', component: SearchCmp},
    {path: 'ui/video', component: VideoViewCmp},
];
@NgModule({
    imports: [RouterModule.forRoot(appRoutes, {})],
    providers: [{provide: APP_BASE_HREF, useValue: ''}],
    exports: [RouterModule]
})
export class AppRoutes {}
