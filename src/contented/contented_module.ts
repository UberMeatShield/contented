import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpClientModule} from '@angular/common/http';
import {MatProgressBarModule} from '@angular/material/progress-bar';
import {MatCardModule} from '@angular/material/card';
import {MatFormFieldModule} from '@angular/material/form-field';
import {MatInputModule} from '@angular/material/input';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';

import {ContentedCmp} from './contented_cmp';
import {SearchCmp} from './search_cmp';
import {DirectoryCmp} from './directory_cmp';
import {ContentedViewCmp} from './contented_view_cmp';
import {ContentedService} from './contented_service';
import {Directory} from './directory';

@NgModule({
  imports: [
      BrowserModule,
      HttpClientModule,
      MatProgressBarModule,
      MatCardModule,
      FormsModule,
      ReactiveFormsModule,
      MatFormFieldModule,
      MatInputModule,
  ],
  declarations: [ContentedCmp, ContentedViewCmp, DirectoryCmp, SearchCmp],
  exports: [ContentedCmp, ContentedViewCmp, SearchCmp],
  providers: [ContentedService]
})
export class ContentedModule {}
