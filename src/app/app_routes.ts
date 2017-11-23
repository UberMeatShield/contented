import {NgModule} from '@angular/core';

import {ContentedCmp} from './../contented/contented_cmp';
import {APP_BASE_HREF} from '@angular/common';
import {RouterModule, Routes} from '@angular/router';

const appRoutes = [
    {path: '', redirectTo: '/ui/main', pathMatch: 'full'},
    {path: 'ui/main', component: ContentedCmp}
];
@NgModule({
    imports: [RouterModule.forRoot(appRoutes)],
    providers: [{provide: APP_BASE_HREF, useValue: ''}],
    exports: [RouterModule]
})
export class AppRoutes {}
