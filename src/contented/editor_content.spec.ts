import { TestBed, ComponentFixture, fakeAsync, tick } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { RouterTestingModule } from '@angular/router/testing';
import { EditorContentCmp } from './editor_content.cmp';
import { DebugElement } from '@angular/core';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { FormControl } from '@angular/forms';
import { MockData } from '../test/mock/mock_data';
import { ContentedModule } from './contented_module';
import { Content } from './content';
import { ApiDef } from './api_def';

declare let $: any;

describe('EditorContentCmp', () => {
  let fixture: ComponentFixture<EditorContentCmp>;
  let cmp: EditorContentCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let httpMock: HttpTestingController;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [RouterTestingModule, HttpClientTestingModule, ContentedModule, NoopAnimationsModule],
      providers: [],
      declarations: [EditorContentCmp],
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

  it('Should be able to render the monaco editor and get a reference', fakeAsync(() => {
    let id = 'A';
    cmp.content = new Content({ id: id, content_type: 'video/mp4' });
    cmp.checkStates = false;
    tick(1000);
    fixture.detectChanges();
    expect($('.vscode-editor-cmp').length).withContext('There should be an editor').toEqual(1);
    tick(1000);

    let url = ApiDef.contented.contentScreens.replace('{mcID}', id);
    httpMock.expectOne(url).flush(MockData.getScreens());
    fixture.detectChanges();
    tick(1000);
    fixture.detectChanges();
    tick(1000);

    let taskUrl = `${ApiDef.tasks.list}?page=1&per_page=100&content_id=${id}`;
    httpMock.expectOne(taskUrl).flush(MockData.taskRequests());
    httpMock.expectOne(r => r.url.includes(ApiDef.contented.tags)).flush(MockData.tags());
    expect($('.screens-form').length).withContext('Video should have the ability to take screens').toEqual(1);
    tick(15000);
    tick(15000);
  }));

  it('Should be able to determine if the content can be video encoded', fakeAsync(() => {
    let content = new Content(MockData.videoContent());
    let vidInfo = content.getVideoInfo();
    let codec = vidInfo.getVideoCodecName();
    expect(codec).toEqual('h264');

    cmp.content = content;
    cmp.checkStates = false;
    fixture.detectChanges();

    let url = ApiDef.contented.contentScreens.replace('{mcID}', cmp.content.id.toString());
    httpMock.expectOne(url).flush(MockData.getScreens());
    fixture.detectChanges();
    tick(10000);

    expect($('.video-encoding-form').length).withContext('It should be a video').toEqual(1);

    let btn = $('.video-encoding-btn');
    expect(btn.length).withContext('We should have an encoding button').toEqual(1);
    expect(btn.attr('disabled')).toEqual(undefined);

    let dupeBtn = $('.duplicate-btn');
    expect(dupeBtn.length).withContext('And be able to search for dupes').toEqual(1);

    let taskUrl = `${ApiDef.tasks.list}?page=1&per_page=100&content_id=${content.id}`;
    httpMock.expectOne(taskUrl).flush(MockData.taskRequests());
    httpMock.expectOne(r => r.url.includes(ApiDef.contented.tags)).flush(MockData.tags());
    tick(15000);
  }));

  it('Should be able to screen new screens', fakeAsync(() => {
    let content = new Content(MockData.videoContent());
    cmp.content = content;
    cmp.checkStates = false;
    fixture.detectChanges();

    let url = ApiDef.contented.contentScreens.replace('{mcID}', cmp.content.id.toString());
    httpMock.expectOne(url).flush(MockData.getScreens());
    fixture.detectChanges();
    tick(10000);

    let taskUrl = `${ApiDef.tasks.list}?page=1&per_page=100&content_id=${content.id}`;
    httpMock.expectOne(taskUrl).flush(MockData.taskRequests());
    httpMock.expectOne(r => r.url.includes(ApiDef.contented.tags)).flush(MockData.tags());
    tick(15000);
  }));
});
