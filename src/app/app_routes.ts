import {NgModule} from '@angular/core';

import {ContentBrowserCmp} from './../contented/content_browser.cmp';
import {VideoBrowserCmp} from './../contented/video_browser.cmp';
import {SearchCmp} from './../contented/search.cmp';

import {ContentViewCmp} from './../contented/content_view.cmp';
import {SplashCmp} from './../contented/splash.cmp';

import {APP_BASE_HREF} from '@angular/common';
import {RouterModule, Routes} from '@angular/router';
import { TasksCmp } from '../contented/tasks.cmp';
import { AdminSearchCmp } from '../contented/admin_search.cmp';
import { EditorContentCmp } from './../contented/editor_content.cmp';

// Hmm, should have made this route have a saner extension
const appRoutes: Routes = [
    {path: '', redirectTo: '/ui/splash', pathMatch: 'full'},
    {path: 'ui/browse/:idx/:rowIdx', component: ContentBrowserCmp, data: {title: 'Browsing Content'},},
    {path: 'ui/video', component: VideoBrowserCmp, data: {title: 'Videos'},},
    {path: 'ui/search', component: SearchCmp, data: {title: 'Search'},},
    {path: 'ui/content/:id', component: ContentViewCmp, data: {title: 'Content View'},},
    {path: 'ui/splash', component: SplashCmp, data: {title: 'Home'},},

    {path: 'admin_ui/editor_content/:id', component: EditorContentCmp, data: {title: 'Edit Content'},},
    {path: 'admin_ui/tasks', component: TasksCmp, data: {title: 'Tasks'},},
    {path: 'admin_ui/search', component: AdminSearchCmp, data: {title: 'Admin Search'},},
];
@NgModule({
    imports: [RouterModule.forRoot(appRoutes, {})],
    providers: [{provide: APP_BASE_HREF, useValue: ''}],
    exports: [RouterModule]
})
export class AppRoutes {}
