import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpClientModule} from '@angular/common/http';
import {MatLegacyProgressBarModule as MatProgressBarModule} from '@angular/material/legacy-progress-bar';
import {MatLegacyCardModule as MatCardModule} from '@angular/material/legacy-card';
import {MatLegacyFormFieldModule as MatFormFieldModule} from '@angular/material/legacy-form-field';
import {MatLegacyInputModule as MatInputModule} from '@angular/material/legacy-input';
import {MatLegacyDialogModule as MatDialogModule} from '@angular/material/legacy-dialog';
import {MatLegacyButtonModule as MatButtonModule} from '@angular/material/legacy-button';
import {MatLegacyPaginatorModule as MatPaginatorModule} from '@angular/material/legacy-paginator';
import {MatRippleModule} from '@angular/material/core';
import {MatLegacyProgressSpinnerModule as MatProgressSpinnerModule} from '@angular/material/legacy-progress-spinner';
import {MatToolbarModule} from '@angular/material/toolbar';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {MatLegacyAutocompleteModule as MatAutocompleteModule} from '@angular/material/legacy-autocomplete';
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
