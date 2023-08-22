import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpClientModule} from '@angular/common/http';
import {MatCardModule as MatCardModule} from '@angular/material/card';
import {MatFormFieldModule as MatFormFieldModule} from '@angular/material/form-field';
import {MatInputModule as MatInputModule} from '@angular/material/input';
import {MatDialogModule as MatDialogModule} from '@angular/material/dialog';
import {MatButtonModule as MatButtonModule} from '@angular/material/button';
import {MatPaginatorModule as MatPaginatorModule} from '@angular/material/paginator';
import {MatRippleModule} from '@angular/material/core';
import {MatProgressSpinnerModule as MatProgressSpinnerModule} from '@angular/material/progress-spinner';
import {MatProgressBarModule as MatProgressBarModule} from '@angular/material/progress-bar';
import {MatToolbarModule} from '@angular/material/toolbar';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {MatAutocompleteModule as MatAutocompleteModule} from '@angular/material/autocomplete';
import {MatIconModule} from '@angular/material/icon';

import {ContentedCmp} from './contented_cmp';
import {ContentedNavCmp} from './contented_nav_cmp';
import {SearchCmp, SearchDialog} from './search_cmp';
import {ContainerCmp} from './container_cmp';
import {ContainerNavCmp} from './container_nav_cmp';
import {ContentedViewCmp} from './contented_view_cmp';
import {ContentViewCmp} from './content_view_cmp';
import {VideoViewCmp, ScreenDialog} from './video_view_cmp';
import {ContentedService} from './contented_service';
import {ScreensCmp} from './screens_cmp';
import {Container} from './container';
import {Content} from './content';
import {Screen} from './screen';
import {ByteFormatterPipe} from './filters';

@NgModule({
  imports: [
      BrowserModule,
      HttpClientModule,
      FormsModule,
      ReactiveFormsModule,
      MatProgressBarModule,
      MatCardModule,
      MatButtonModule,
      MatDialogModule,
      MatFormFieldModule,
      MatInputModule,
      MatIconModule,
      MatPaginatorModule,
      MatRippleModule,
      MatProgressSpinnerModule,
      MatAutocompleteModule,
      MatToolbarModule,
  ],
  declarations: [
      ContentedCmp,
      ContentedNavCmp,
      ContentedViewCmp,
      ContainerCmp,
      ContainerNavCmp,
      ContentViewCmp,
      VideoViewCmp,
      ScreenDialog,
      SearchCmp,
      SearchDialog,
      ScreensCmp,
      ByteFormatterPipe
  ],
  exports: [
      ContentedCmp,
      ContentedNavCmp,
      ContentedViewCmp,
      ContainerCmp,
      ContainerNavCmp,
      SearchCmp,
      VideoViewCmp,
      ContentViewCmp,
      SearchDialog,
  ],
  providers: [ContentedService]
})
export class ContentedModule {}
