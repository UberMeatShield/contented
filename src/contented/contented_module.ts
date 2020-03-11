import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpClientModule} from '@angular/common/http';
import {MatProgressBarModule} from '@angular/material/progress-bar';
import {FormsModule} from '@angular/forms';

import {ContentedCmp} from './contented_cmp';
import {DirectoryCmp} from './directory_cmp';
import {ContentedViewCmp} from './contented_view_cmp';
import {ContentedService} from './contented_service';
import {Directory} from './directory';

@NgModule({
  imports: [BrowserModule, HttpClientModule, MatProgressBarModule, FormsModule],
  declarations: [ContentedCmp, ContentedViewCmp, DirectoryCmp],
  exports: [ContentedCmp, ContentedViewCmp],
  providers: [ContentedService]
})
export class ContentedModule {}
