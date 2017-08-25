import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpModule} from '@angular/http';
import {AppRoutes} from './app_routes';
import {App} from './app';

@NgModule({
  imports: [BrowserModule, AppRoutes],
  declarations: [App],
  bootstrap: [App]
})
export class AppModule {}
