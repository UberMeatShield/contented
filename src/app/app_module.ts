import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { AppRoutes } from './app_routes';
import { ContentedModule } from './../contented/contented_module';
import { App } from './app';

import { environment } from './../environments/environment';
import { BrowserAnimationsModule, NoopAnimationsModule } from '@angular/platform-browser/animations';
let AnimationsModule = !environment.production ? NoopAnimationsModule : BrowserAnimationsModule;

@NgModule({ declarations: [App],
    bootstrap: [App], imports: [BrowserModule, AppRoutes, AppRoutes, ContentedModule, AnimationsModule], providers: [provideHttpClient(withInterceptorsFromDi())] })
export class AppModule {}
