import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpModule} from '@angular/http';
import {AppRoutes} from './app_routes';
import {ContentedCmp} from './contented_cmp';
import {App} from './app';

@NgModule({
  imports: [BrowserModule, AppRoutes, HttpModule, AppRoutes],
  declarations: [App, ContentedCmp],
  exports: [ContentedCmp],
  bootstrap: [App]
})
export class AppModule {}
