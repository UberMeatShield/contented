import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpClientModule} from '@angular/common/http';
import {AppRoutes} from './app_routes';
import {ContentedModule} from './../contented/contented_module';
import {App} from './app';


@NgModule({
  imports: [BrowserModule, AppRoutes, HttpClientModule, AppRoutes, ContentedModule],
  declarations: [App],
  bootstrap: [App]
})
export class AppModule {}
