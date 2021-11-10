import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpClientModule} from '@angular/common/http';
import {FormsModule} from '@angular/forms';
import {AppRoutes} from './app_routes';
import {ContentedModule} from './../contented/contented_module';
import {App} from './app';

import {environment} from './../environments/environment';
import {BrowserAnimationsModule, NoopAnimationsModule} from '@angular/platform-browser/animations';
let AnimationsModule = environment['test'] ? NoopAnimationsModule : BrowserAnimationsModule;

@NgModule({
  imports: [
      BrowserModule,
      AppRoutes,
      HttpClientModule,
      AppRoutes,
      ContentedModule,
      AnimationsModule,
  ],
  declarations: [App],
  bootstrap: [App]
})
export class AppModule {}
