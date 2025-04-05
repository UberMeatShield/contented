import { TestBed, ComponentFixture, waitForAsync, fakeAsync, tick, flush } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { MockData } from '../test/mock/mock_data';
import { MonacoLoaded, WaitForMonacoLoad, ContentedModule } from './contented_module';
import { TagLang, TAGS_RESPONSE } from './tagging_syntax';
import { ApiDef } from './api_def';
import { VSCodeEditorCmp } from './vscode_editor.cmp';

declare let $: any;
let editorValue = ` class Funky() {
  public answer: number = 42;
   constructor (zug: number) {
   }
   monkey() { . Wagggh
   }

   Google Earth

   what the heck
}`;
import _ from 'lodash';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('VSCodeEditorCmp', () => {
  let fixture: ComponentFixture<VSCodeEditorCmp>;
  let cmp: VSCodeEditorCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let httpMock: HttpTestingController;
  let tagLang: TagLang;

  beforeEach(() => {
    TestBed.configureTestingModule({
    declarations: [VSCodeEditorCmp],
    teardown: { destroyAfterEach: true },
    imports: [ContentedModule, NoopAnimationsModule],
    providers: [provideHttpClient(withInterceptorsFromDi()), provideHttpClientTesting()]
}).compileComponents();

    fixture = TestBed.createComponent(VSCodeEditorCmp);
    de = fixture.debugElement.query(By.css('.vscode-editor-cmp'));
    el = de.nativeElement;
    cmp = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);

    tagLang = new TagLang();
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('should create the app', () => {
    expect(cmp).toBeTruthy();
    expect(el).toBeTruthy();
  });

  // TODO: The ajax load of tags is still not working right.
  it('Should be able to render the monaco editor and process tokens', waitForAsync(async () => {
    cmp.language = 'test';
    cmp.editorValue = editorValue;

    let keywords = [
      { id: 'constructor', tag_type: 'keywords' },
      { id: 'number', tag_type: 'keywords' },
      { id: 'class', tag_type: 'keywords' },
      { id: 'Google Earth', tag_type: 'typeKeywords' },
      { id: 'Wagggh', tag_type: 'typeKeywords' },
    ];
    let tags = {
      total: 4,
      results: keywords,
    };

    TAGS_RESPONSE.initialized = false;

    cmp.tags = [];
    fixture.detectChanges();
    await WaitForMonacoLoad();
    await cmp.isInitialized();
    await fixture.whenRenderingDone();
    expect(cmp.initialized).toBe(true);

    httpMock.expectOne(r => r.url.includes(ApiDef.contented.tags)).flush(tags);
    expect(cmp.problemTags.length).withContext('We should consider Google Earth a problem').toBe(1);
    fixture.detectChanges();

    expect((window as any).monaco?.languages).toBeDefined();
    let ids = _.map(keywords, 'id').slice(0, keywords.length - 2);
    let types = _.map(keywords, 'id').slice(keywords.length - 2, keywords.length);
    tagLang.setMonacoLanguage(cmp.language, ids, types);

    expect(cmp.descriptionControl?.value).toEqual(editorValue);
    let tokens = cmp.getTokens();

    expect(cmp.editor).withContext('It should have initialized').toBeDefined();
    expect(tokens.sort()).toEqual(ids.sort());
    expect(cmp.monacoEditor).toBeDefined();

    // Qoutes in the string can still be a problem
    let tokenTypes = cmp.getTokens('type');
    expect(tokenTypes.sort()).toEqual(types);
    console.log('End test case');
  }));
});
