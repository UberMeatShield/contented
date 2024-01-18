import {NgModule} from '@angular/core';

import {ContentBrowserCmp} from './../contented/content_browser.cmp';
import {VideoBrowserCmp} from './../contented/video_browser.cmp';
import {SearchCmp} from './../contented/search.cmp';

import {EditorContentCmp} from './../contented/editor_content.cmp';
import {ContentViewCmp} from './../contented/content_view.cmp';
import {SplashCmp} from './../contented/splash.cmp';

import {APP_BASE_HREF} from '@angular/common';
import {RouterModule, Routes} from '@angular/router';
import { TasksCmp } from '../contented/tasks.cmp';

// Hmm, should have made this route have a saner extension
const appRoutes: Routes = [
    {path: '', redirectTo: '/ui/splash', pathMatch: 'full'},
    {path: 'ui/browse/:idx/:rowIdx', component: ContentBrowserCmp},
    {path: 'ui/video', component: VideoBrowserCmp},
    {path: 'ui/search', component: SearchCmp},
    {path: 'ui/content/:id', component: ContentViewCmp},
    {path: 'ui/editor_content/:id', component: EditorContentCmp},
    {path: 'ui/splash', component: SplashCmp},
    {path: 'admin_ui/tasks', component: TasksCmp},
];
@NgModule({
    imports: [RouterModule.forRoot(appRoutes, {})],
    providers: [{provide: APP_BASE_HREF, useValue: ''}],
    exports: [RouterModule]
})
export class AppRoutes {}
