import { TestBed, ComponentFixture, waitForAsync, fakeAsync, tick} from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { RouterTestingModule } from '@angular/router/testing';
import { DebugElement } from '@angular/core';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import {NoopAnimationsModule} from '@angular/platform-browser/animations';
import {MockData} from '../test/mock/mock_data';
import {MonacoLoaded, WaitForMonacoLoad, ContentedModule} from './contented_module';
import {TagLang} from './tagging_syntax';
import {ApiDef} from './api_def';
import { VSCodeEditorCmp } from './vscode_editor.cmp';

declare let $ : any;
let editorValue = ` class Funky() {
  public answer: number = 42;
   constructor (zug: number) {
   }
   monkey() {
   }

   Google Earth

   what the heck
}`;
import * as _ from 'lodash-es';

describe('VSCodeEditorCmp', () => {

  let fixture: ComponentFixture<VSCodeEditorCmp>;
  let cmp: VSCodeEditorCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let httpMock: HttpTestingController;
  let tagLang: TagLang;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [
        RouterTestingModule,
        HttpClientTestingModule,
        ContentedModule,
        NoopAnimationsModule,
      ],
      providers: [
      ],
      declarations: [
        VSCodeEditorCmp
      ],
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
  fit("Should be able to render the monaco editor and process tokens", waitForAsync(() => {
    cmp.language = "test";
    cmp.editorValue = editorValue;
    fixture.detectChanges()

    let keywords = [
        {id: "constructor", tag_type: "keywords"},
        {id: "number", tag_type: "keywords"},
        {id: "class", tag_type: "keywords"},
        {id: "Google Earth", tag_type: "typeKeywords"},
    ]
    let tags = {
      total: 4,
      results: keywords
    }

    WaitForMonacoLoad();
    httpMock.expectOne(r => r.url.includes(ApiDef.contented.tags)).flush(tags);
    expect(cmp.problemTags.length).withContext("We should consider Google Earth a problem").toBe(1);

    fixture.whenStable().then(() => {
      expect((window as any).monaco).toBeDefined()
      let ids = _.map(keywords, 'id')
      tagLang.setMonacoLanguage(cmp.language, ids.slice(0, 3), ["Google Earth", "WTF Mate"]);

      expect(cmp.descriptionControl.value).toEqual(editorValue);
      let tokens = cmp.getTokens();
      let tokenType = cmp.getTokens("type");
      console.log("tokenType", tokenType);

      expect(cmp.editor).withContext("It should have initialized").toBeDefined();
      expect(tokens.sort()).toEqual(ids.sort())
      expect(cmp.monacoEditor).toBeDefined();
    })
  }));
});