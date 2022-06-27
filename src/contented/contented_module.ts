import {NgModule} from '@angular/core';
import {BrowserModule} from '@angular/platform-browser';
import {HttpClientModule} from '@angular/common/http';
import {MatProgressBarModule} from '@angular/material/progress-bar';
import {MatCardModule} from '@angular/material/card';
import {MatFormFieldModule} from '@angular/material/form-field';
import {MatInputModule} from '@angular/material/input';
import {MatDialogModule} from '@angular/material/dialog';
import {MatButtonModule} from '@angular/material/button';
import {MatPaginatorModule} from '@angular/material/paginator';
import {MatRippleModule} from '@angular/material/core';
import {MatProgressSpinnerModule} from '@angular/material/progress-spinner';
import {MatToolbarModule} from '@angular/material/toolbar';
import {FormsModule, ReactiveFormsModule} from '@angular/forms';
import {MatAutocompleteModule} from '@angular/material/autocomplete';

import {ContentedCmp} from './contented_cmp';
import {ContentedNavCmp} from './contented_nav_cmp';
import {SearchCmp, SearchDialog} from './search_cmp';
import {ContainerCmp} from './container_cmp';
import {ContainerNavCmp} from './container_nav_cmp';
import {ContentedViewCmp} from './contented_view_cmp';
import {MediaViewCmp} from './media_view_cmp';
import {VideoViewCmp} from './video_view_cmp';
import {ContentedService} from './contented_service';
import {ScreensCmp} from './screens_cmp';
import {Container} from './container';
import {Media} from './media';
import {Screen} from './screen';

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
      MediaViewCmp,
      VideoViewCmp,
      SearchCmp,
      SearchDialog,
      ScreensCmp,
  ],
  exports: [
      ContentedCmp,
      ContentedNavCmp,
      ContentedViewCmp,
      ContainerCmp,
      ContainerNavCmp,
      SearchCmp,
      VideoViewCmp,
      MediaViewCmp,
      SearchDialog,
  ],
  providers: [ContentedService]
})
export class ContentedModule {}
