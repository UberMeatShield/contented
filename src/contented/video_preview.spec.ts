import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpParams } from '@angular/common/http';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import { ContentedService } from './contented_service';
import { ContentedModule } from './contented_module';
import { VideoPreviewCmp } from './video_preview.cmp';
import { Content } from './content';

import * as _ from 'lodash';
import { MockData } from '../test/mock/mock_data';
import { ApiDef } from './api_def';

declare var $;
describe('TestingVideoPreviewCmp', () => {
  let fixture: ComponentFixture<VideoPreviewCmp>;
  let service: ContentedService;
  let comp: VideoPreviewCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let router: Router;

  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [
        RouterTestingModule.withRoutes([{ path: 'screens/:screenId', component: VideoPreviewCmp }]),
        FormsModule,
        ContentedModule,
        HttpClientTestingModule,
      ],
      providers: [ContentedService],
    }).compileComponents();

    service = TestBed.get(ContentedService);
    fixture = TestBed.createComponent(VideoPreviewCmp);
    httpMock = TestBed.get(HttpTestingController);
    comp = fixture.componentInstance;
    de = fixture.debugElement.query(By.css('.video-preview-cmp'));
    el = de.nativeElement;
    router = TestBed.get(Router);
    router.initialNavigation();
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a screens view component', () => {
    const info = MockData.videoContent();
    info.preview = "";
    let content = new Content(info);

    expect(content.isVideo()).toBe(true);

    expect(content.shouldUseTypedPreview()).toEqual('videocam');
    expect(content.videoInfo).toBeDefined();
    expect(content.videoInfo?.format?.duration).toEqual(10);
  });

  it('Should initialize the video preview component', () => {
    expect(el).toBeDefined();
    expect(comp).toBeDefined();
  });

  it('Should load video content when given a content ID', fakeAsync(() => {
    comp.content = new Content(MockData.videoContent());
    fixture.detectChanges();
    expect($('.video-view-card').length).toEqual(1);

    let url = ApiDef.contented.contentScreens.replace('{mcID}', comp.content.id.toString());
    let req = httpMock.expectOne(req => req.url === url);
    let sRes = MockData.getScreens();
    req.flush(sRes);
    tick(1000);
    fixture.detectChanges();

    expect($('.screen').length).toEqual(sRes.results.length);
    expect($('.video-duration').text()).toEqual('Duration: 00:00:10');
  }));
});
