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
import {FormsModule, ReactiveFormsModule} from '@angular/forms';

import {ContentedCmp} from './contented_cmp';
import {ContentedNavCmp} from './contented_nav_cmp';
import {SearchCmp, SearchDialog} from './search_cmp';
import {ContainerCmp} from './container_cmp';
import {ContainerNavCmp} from './container_nav_cmp';
import {ContentedViewCmp} from './contented_view_cmp';
import {ContentedService} from './contented_service';
import {Container} from './container';
import {Media} from './media';

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
  ],
  declarations: [ContentedCmp, ContentedNavCmp, ContentedViewCmp, ContainerCmp, ContainerNavCmp, SearchCmp, SearchDialog],
  exports: [ContentedCmp, ContentedNavCmp, ContentedViewCmp, ContainerCmp, ContainerNavCmp, SearchCmp, SearchDialog],
  providers: [ContentedService]
})
export class ContentedModule {}
