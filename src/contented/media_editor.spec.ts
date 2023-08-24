import { TestBed, ComponentFixture, fakeAsync, tick} from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { RouterTestingModule } from '@angular/router/testing';
import { MediaEditorCmp } from './media_editor.component';
import { DebugElement } from '@angular/core';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import {FormControl} from '@angular/forms';
import {MockData} from '../test/mock/mock_data';
import {ContentedModule} from './contented_module';
import {Content} from './content';
import {ApiDef} from './api_def';

declare let $ : any;

describe('MediaEditorCmp', () => {

  let fixture: ComponentFixture<MediaEditorCmp>;
  let cmp: MediaEditorCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [
        RouterTestingModule,
        HttpClientTestingModule,
        ContentedModule,
      ],
      providers: [
      ],
      declarations: [
        MediaEditorCmp
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(MediaEditorCmp);
    de = fixture.debugElement.query(By.css('.media-editor-cmp'));
    el = de.nativeElement;
    cmp = fixture.componentInstance;
    httpMock = TestBed.inject(HttpTestingController);
  });

  afterEach(() => {
      httpMock.verify();
  });

  it('should create the app', () => {
    expect(cmp).toBeTruthy();
    expect(el).toBeTruthy();
  });

  it("Should be able to render the monaco editor and get a reference", fakeAsync(() => {
    cmp.readOnly = true;
    cmp.mc = new Content({id: 'A', content_type: 'video/mp4'});
    tick(1000);
    fixture.detectChanges();
    expect(cmp.editor).withContext("Monaco should init and have a reference").toBeDefined();
    cmp.setReadOnly(true);
    tick(1000);

    let url = ApiDef.contented.contentScreens.replace("{mcID}", cmp.mc.id);
    httpMock.expectOne(url).flush(MockData.getScreens());
    fixture.detectChanges();
  }));
});

