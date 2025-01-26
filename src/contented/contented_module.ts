import { NgModule } from '@angular/core';
import { BrowserModule, Title } from '@angular/platform-browser';
import { HttpClientModule } from '@angular/common/http';
import { RouterModule } from '@angular/router';

import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatDialogModule } from '@angular/material/dialog';
import { MatButtonModule } from '@angular/material/button';
import { MatPaginatorModule } from '@angular/material/paginator';
import { MatRippleModule } from '@angular/material/core';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatToolbarModule } from '@angular/material/toolbar';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import { MatAutocompleteModule as MatAutocompleteModule } from '@angular/material/autocomplete';
import { MatIconModule } from '@angular/material/icon';
import { MatTableModule } from '@angular/material/table';
import { MatSelectModule } from '@angular/material/select';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { MatMenuModule } from '@angular/material/menu';
import { MatRadioModule } from '@angular/material/radio';

import { ContentBrowserCmp } from './content_browser.cmp';
import { ContentedNavCmp } from './contented_nav.cmp';
import { SearchCmp, SearchDialog } from './search.cmp';
import { AdminSearchCmp } from './admin_search.cmp';
import { AdminContainersCmp } from './admin_containers.cmp';
import { ContainerCmp } from './container.cmp';
import { ContainerNavCmp } from './container_nav.cmp';
import { ContentedViewCmp } from './contented_view_cmp';
import { ContentViewCmp } from './content_view.cmp';
import { VideoBrowserCmp } from './video_browser.cmp';
import { VideoPreviewCmp, ScreenDialog } from './video_preview.cmp';
import { ContentedService } from './contented_service';
import { ScreensCmp } from './screens.cmp';
import { EditorContentCmp } from './editor_content.cmp';
import { SplashCmp } from './splash.cmp';
import { TagsCmp } from './tags.cmp';
import { VSCodeEditorCmp } from './vscode_editor.cmp';
import { TaskRequestCmp } from './taskrequest.cmp';
import { TasksCmp } from './tasks.cmp';
import { ErrorHandlerCmp, ErrorDialogCmp } from './error_handler.cmp';
import { ByteFormatterPipe, DurationFormatPipe } from './filters';
import { SafePipe } from './safe.pipe';

import { MonacoEditorModule, NgxMonacoEditorConfig } from 'ngx-monaco-editor-v2';
import { TagLang } from './tagging_syntax';

import { FavoritesCmp } from './favorites.cmp';
import { PreviewContentCmp } from './preview_content.cmp';

let MONACO_LOADED = false;
let GIVE_UP = 0;
function monacoPoller(resolve, reject) {
  if (MONACO_LOADED) {
    return resolve((window as any).monaco);
  } else {
    if (GIVE_UP > 4) {
      reject('Monaco was not loaded within the required timeframe');
    }
    GIVE_UP++;
    setTimeout(() => {
      monacoPoller(resolve, reject);
    }, GIVE_UP * 500);
  }
}

export let MonacoLoaded: Promise<any>;

export async function WaitForMonacoLoad() {
  MonacoLoaded = new Promise((resolve, reject) => {
    return monacoPoller(resolve, reject);
  });
  return await MonacoLoaded.then(() => {
    console.log('Monaco Resolved this is used in test Cases');
  });
}

// Kinda annoying this has to be configured like this but I suppose it ok.
const monacoConfig: NgxMonacoEditorConfig = {
  baseUrl: '/public/static',
  defaultOptions: {
    wordWrap: 'on',
    minimap: { enabled: false },
    scrollbar: {
      alwaysConsumeMouseWheel: false, // This prevents from intercepting the page scroll
      handleMouseWheel: false, // This prevents it from scrolling and hiding editing
    },
  },
  onMonacoLoad: () => {
    // Can just make this do a call to the system and pull back a file that is generated that has the tags.
    /*
    console.log("Now here is where we register a new language for tags.");
    let monaco = (<any>window).monaco;
    let lang = monaco.languages;
    lang.register({id: "tagging"});
    lang.setMonarchTokensProvider("tagging", TAGGING_SYNTAX);
    */
    console.log('onMonacoLoad has been called');
    MONACO_LOADED = true;
    let tl = new TagLang();
    tl.loadLanguage((<any>window).monaco, 'tagging');
  },
};

@NgModule({
  imports: [
    BrowserModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,
    RouterModule,
    MonacoEditorModule.forRoot(monacoConfig),
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
    MatTableModule,
    MatSelectModule,
    MatSnackBarModule,
    MatMenuModule,
    MatRadioModule,
  ],
  declarations: [
    ContentBrowserCmp,
    ContentedNavCmp,
    ContentedViewCmp,
    ContainerCmp,
    ContainerNavCmp,
    ContentViewCmp,
    VideoBrowserCmp,
    VideoPreviewCmp,
    ScreenDialog,
    AdminSearchCmp,
    AdminContainersCmp,
    SearchCmp,
    SearchDialog,
    ScreensCmp,
    SplashCmp,
    FavoritesCmp,
    ByteFormatterPipe,
    DurationFormatPipe,
    SafePipe,
    EditorContentCmp,
    PreviewContentCmp,
    TagsCmp,
    TaskRequestCmp,
    TasksCmp,
    VSCodeEditorCmp,
    ErrorHandlerCmp,
    ErrorDialogCmp,
  ],
  exports: [
    ContentBrowserCmp,
    ContentedNavCmp,
    ContentedViewCmp,
    ContainerCmp,
    ContainerNavCmp,
    AdminSearchCmp,
    AdminContainersCmp,
    SearchCmp,
    VideoBrowserCmp,
    ContentViewCmp,
    SearchDialog,
    TagsCmp,
    TaskRequestCmp,
    TasksCmp,
    VSCodeEditorCmp,
    ErrorHandlerCmp,
  ],
  providers: [ContentedService, Title],
})
export class ContentedModule {}
