import { fakeAsync, getTestBed, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpParams, provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import { ContentedViewCmp } from '../contented/contented_view_cmp';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';
import { Content } from '../contented/content';
import { Container } from '../contented/container';
import { GlobalNavEvents } from '../contented/nav_events';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import * as _ from 'lodash';
import { MockData } from '../test/mock/mock_data';

declare var $;
describe('TestingContentedViewCmp', () => {
  let fixture: ComponentFixture<ContentedViewCmp>;
  let service: ContentedService;
  let comp: ContentedViewCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let router: Router;

  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
    teardown: { destroyAfterEach: false },
    imports: [RouterTestingModule.withRoutes([{ path: 'ui/:idx/:rowIdx', component: ContentedViewCmp }]),
        FormsModule,
        ContentedModule,
        NoopAnimationsModule],
    providers: [ContentedService, provideHttpClient(withInterceptorsFromDi()), provideHttpClientTesting()]
}).compileComponents();

    service = TestBed.get(ContentedService);
    fixture = TestBed.createComponent(ContentedViewCmp);
    httpMock = TestBed.get(HttpTestingController);
    comp = fixture.componentInstance;

    de = fixture.debugElement.query(By.css('.contented-view-cmp'));
    el = de.nativeElement;
    router = TestBed.get(Router);
    router.initialNavigation();
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a contented component', () => {
    expect(comp).toBeDefined('We should have the Contented comp');
    expect(el).toBeDefined('We should have a top level element');
  });

  it('Can render an image and render', () => {
    comp.content = null;
    comp.visible = true;
    fixture.detectChanges();
    expect($('.content-full-view').length).toBe(0, 'It should not be visible');

    let img = MockData.getImg();
    comp.content = img;
    fixture.detectChanges();
    expect($('.content-full-view').length).toBe(1, 'It should be visible');
  });

  it('Forcing a width and height will be respected', () => {
    comp.content = MockData.getImg();
    comp.forceWidth = 666;
    comp.forceHeight = 42;
    comp.visible = true;
    fixture.detectChanges();
    expect($('.content-full-view').length).toBe(1, 'It should be visible');

    window.dispatchEvent(new Event('resize'));
    fixture.detectChanges();
    // It should be forcing a detection of the resize (otherwise it is calculated already)
    // comp.calculateDimensions();

    expect(comp.maxWidth).toEqual(comp.forceWidth, 'Ensure width assignment works');
    expect(comp.maxHeight).toEqual(comp.forceHeight, 'Ensure height assignment works');
  });

  // Test that we listen to nav events correctly
  it('Should register nav events', fakeAsync(() => {
    fixture.detectChanges();
    expect($('.content-full-view').length).toBe(0, 'Nothing in the view');

    let initialSel = new Content({ id: 23, content_type: 'image/png' });
    GlobalNavEvents.selectContent(initialSel, new Container({ id: 1 }));
    fixture.detectChanges();
    expect(comp.content).toEqual(initialSel);

    let content = MockData.getImg();
    expect(content.content_type).toEqual('image/png');
    GlobalNavEvents.viewFullScreen(content);
    tick(1000);
    fixture.detectChanges();
    expect($('.content-full-view').length).toBe(1, 'It should now be visible');

    expect(comp.content).toEqual(content);
    expect(comp.visible).toBeTrue();
    expect($('.full-view-img').length).withContext('It should be an image').toEqual(1);
    expect(comp.content).toEqual(content, 'A view event with a content item should change it');

    GlobalNavEvents.hideFullScreen();
    fixture.detectChanges();
    expect(comp.visible).toBe(false, 'It should not be visible now');
  }));

  it('Should have a video in the case of a video, image for image', () => {
    let video = new Content({ content_type: 'video/mp4', fullUrl: 'cthulhu' });
    let img = new Content({
      content_type: 'image/jpeg',
      fullUrl: 'cat/pics.jpg',
    });

    comp.visible = true;
    comp.content = img;
    fixture.detectChanges();
    expect($('.content-video-screens').length).toEqual(0, 'No screens');
    expect($('img').length).toEqual(1, 'We should have an image');
    expect($('video').length).toEqual(0, 'Not a video');

    comp.content = video;
    fixture.detectChanges();
    expect($('video').length).toEqual(1, 'Now it should be a video');
    expect($('image').length).toEqual(0, 'Not an image');
    expect(video.isVideo()).toBe(true, 'It should be a video file');
    fixture.detectChanges();
    expect($('.content-video-screens').length).toEqual(1, 'It should have screens');
  });
});
