import {NgModule} from '@angular/core';

// Rename this to browse
import {ContentBrowserCmp} from './../contented/content_browser.cmp';
import {VideoViewCmp} from './../contented/video_view.cmp';
import {SearchCmp} from './../contented/search.cmp';

// Rename media to Content
import {MediaEditorCmp} from './../contented/media_editor.component';
import {ContentViewCmp} from './../contented/content_view.cmp';

import {APP_BASE_HREF} from '@angular/common';
import {RouterModule, Routes} from '@angular/router';

// Hmm, should have made this route have a saner extension
const appRoutes: Routes = [
    {path: '', redirectTo: '/ui/browse/0/0', pathMatch: 'full'},
    {path: 'ui/browse/:idx/:rowIdx', component: ContentBrowserCmp},
    {path: 'ui/content/:id', component: ContentViewCmp},
    {path: 'ui/search', component: SearchCmp},
    {path: 'ui/video', component: VideoViewCmp},
    {path: 'ui/media_editor/:id', component: MediaEditorCmp},
];
@NgModule({
    imports: [RouterModule.forRoot(appRoutes, {})],
    providers: [{provide: APP_BASE_HREF, useValue: ''}],
    exports: [RouterModule]
})
export class AppRoutes {}
