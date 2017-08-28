import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpModule} from '@angular/http';

import {ContentedCmp} from './contented_cmp';
import {ContentedService} from './contented_service';

@NgModule({
  imports: [BrowserModule, HttpModule],
  declarations: [ContentedCmp],
  exports: [ContentedCmp],
  providers: [ContentedService]
})
export class ContentedModule {}
