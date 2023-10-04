import { TestBed, ComponentFixture, fakeAsync, tick} from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { RouterTestingModule } from '@angular/router/testing';
import { EditorContentCmp } from './editor_content.cmp';
import { DebugElement } from '@angular/core';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import {NoopAnimationsModule} from '@angular/platform-browser/animations';
import {FormControl} from '@angular/forms';
import {MockData} from '../test/mock/mock_data';
import {ContentedModule} from './contented_module';
import {Content} from './content';
import {ApiDef} from './api_def';

declare let $ : any;

describe('EditorContentCmp', () => {

  let fixture: ComponentFixture<EditorContentCmp>;
  let cmp: EditorContentCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let httpMock: HttpTestingController;

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
        EditorContentCmp
      ],
    }).compileComponents();

    fixture = TestBed.createComponent(EditorContentCmp);
    de = fixture.debugElement.query(By.css('.editor-content-cmp'));
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
    cmp.content = new Content({id: 'A', content_type: 'video/mp4'});
    tick(1000);
    fixture.detectChanges();
    expect($(".vscode-editor-cmp").length).withContext("There should be an editor").toEqual(1);
    tick(1000);

    let url = ApiDef.contented.contentScreens.replace("{mcID}", cmp.content.id);
    httpMock.expectOne(url).flush(MockData.getScreens());
    fixture.detectChanges();
    tick(1000);
    fixture.detectChanges();
    tick(1000);

    expect($(".screens-form").length).withContext("Video should have the ability to take screens").toEqual(1);
  }));

  it("Should be able to determine if the content can be video encoded", fakeAsync(() => {
    let content = new Content(MockData.videoContent());
    let vidInfo = content.getVideoInfo();
    let codec = vidInfo.getVideoCodecName();
    expect(codec).toEqual("h264")

    cmp.content = content;
    fixture.detectChanges();

    let url = ApiDef.contented.contentScreens.replace("{mcID}", cmp.content.id);
    httpMock.expectOne(url).flush(MockData.getScreens());
    fixture.detectChanges();
    tick(10000);

    expect($(".video-encoding-form").length).withContext("It should be a video").toEqual(1);

    let btn = $(".video-encoding-btn");
    expect(btn.length).withContext("We should have an encoding button").toEqual(1);
    expect(btn.attr("disabled")).toEqual(undefined);
  }));
});

