import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';

import { ContentedModule } from '../contented/contented_module';
import { FavoritesCmp } from './favorites.cmp';
import { MockData } from '../test/mock/mock_data';

import $ from 'jquery';
import { GlobalNavEvents } from './nav_events';
import { getFavorites } from './container';
import { ApiDef } from './api_def';
import { Content, Tag } from './content';
import { z } from 'zod';

describe('FavoritesCmp', () => {
  let comp: FavoritesCmp;
  let fixture: ComponentFixture<FavoritesCmp>;

  let httpMock: HttpTestingController;
  let el: HTMLElement;
  let de: DebugElement;

  beforeEach(() => {
    TestBed.configureTestingModule({
      imports: [NoopAnimationsModule, ContentedModule, HttpClientTestingModule],
      declarations: [FavoritesCmp],
    });
    fixture = TestBed.createComponent(FavoritesCmp);
    comp = fixture.componentInstance;

    httpMock = TestBed.inject(HttpTestingController);
    de = fixture.debugElement.query(By.css('.favorites-cmp'));
    el = de?.nativeElement;

    const favorites = getFavorites();
    favorites.contents = [];
    favorites.total = 0;
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('Should be able to Initialize the component', fakeAsync(() => {
    expect(comp).toBeDefined();
    expect(el).toBeDefined();
  }));

  it('Should be able to render a favorite', fakeAsync(() => {
    fixture.detectChanges();
    expect($('.preview-content-cmp').length).toBe(0);

    let contentData = MockData.getImg();
    const content = new Content(contentData);

    GlobalNavEvents.favoriteContent(MockData.getImg());
    tick(100);
    fixture.detectChanges();
    expect($('.preview-content-cmp').length).toBe(1);
  }));

  it('Should be able to remove a favorite video', fakeAsync(() => {
    const content = MockData.getVideo();


    getFavorites().addContents([content]);
    comp.container = getFavorites();

    fixture.detectChanges();
    expect($('.preview-content-cmp').length).toBe(1);

    GlobalNavEvents.favoriteContent(MockData.getVideo());
    tick(100);
    fixture.detectChanges();
    expect($('.preview-content-cmp').length).toBe(0);
  }));

  it('Should be able to toggle duplicate', fakeAsync(() => {
    const content = MockData.getVideo();
    getFavorites().addContents([content]);
    comp.container = getFavorites();

    fixture.detectChanges();

    comp.toggleDuplicate(content);
    httpMock.expectOne({
      method: 'PUT',
      url: ApiDef.contented.content.replace('{id}', content.id.toString()),
    });

    tick(100);
    fixture.detectChanges();
    expect(content.duplicate).toBe(true);
  }));
});
