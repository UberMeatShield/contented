import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';

import { ContentedModule } from '../contented/contented_module';
import { PreviewContentCmp } from './preview_content.cmp';
import { MockData } from '../test/mock/mock_data';

import $ from 'jquery';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('PreviewContentCmp', () => {
  let comp: PreviewContentCmp;
  let fixture: ComponentFixture<PreviewContentCmp>;

  let httpMock: HttpTestingController;
  let el: HTMLElement;
  let de: DebugElement;

  beforeEach(() => {
    TestBed.configureTestingModule({
    declarations: [PreviewContentCmp],
    imports: [NoopAnimationsModule, ContentedModule],
    providers: [provideHttpClient(withInterceptorsFromDi()), provideHttpClientTesting()]
});
    fixture = TestBed.createComponent(PreviewContentCmp);
    comp = fixture.componentInstance;

    httpMock = TestBed.inject(HttpTestingController);
    de = fixture.debugElement.query(By.css('.preview-content-cmp'));
    el = de?.nativeElement;
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('Should be trying to load tasks', fakeAsync(() => {
    expect(comp).toBeDefined();
    expect(el).toBeDefined();
  }));

  it('Should have a preview-content-cmp class', fakeAsync(() => {
    comp.content = MockData.getImg();
    fixture.detectChanges();
    expect($('.preview-is').length).toBe(1);
    expect($('.preview-img').length).toBe(1);
  }));

  it('should be able to handle a video', fakeAsync(() => {
    const content = MockData.getVideo();
    comp.content = content;
    expect(content.isVideo()).toBe(true);
    fixture.detectChanges();
    expect($('.video-overlay').length).toBe(1);
  }));
});
