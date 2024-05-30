import { fakeAsync, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpClientTestingModule, HttpTestingController } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import { SplashCmp } from '../contented/splash.cmp';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';
import { ApiDef } from '../contented/api_def';

import * as _ from 'lodash';
import { MockData } from '../test/mock/mock_data';

declare var $;

describe('TestingSplashCmp', () => {
  let fixture: ComponentFixture<SplashCmp>;
  let service: ContentedService;
  let comp: SplashCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let router: Router;

  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
      imports: [
        RouterTestingModule.withRoutes([{ path: 'ui/view/:id', component: SplashCmp }]),
        FormsModule,
        ContentedModule,
        HttpClientTestingModule,
      ],
      providers: [ContentedService],
    }).compileComponents();

    service = TestBed.inject(ContentedService);
    fixture = TestBed.createComponent(SplashCmp);
    httpMock = TestBed.inject(HttpTestingController);
    comp = fixture.componentInstance;

    de = fixture.debugElement.query(By.css('.splash-cmp'));
    el = de.nativeElement;
    router = TestBed.get(Router);
    router.initialNavigation();
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a contented component', () => {
    expect(comp).withContext('We should have the SplashCmp').toBeDefined();
    expect(el).withContext('We should have a top level element').toBeDefined();
  });

  it('Fully handles routing arguments', fakeAsync(() => {
    // Loads content (splash call)
    fixture.detectChanges();
    tick(10000);

    let splash = MockData.splash();
    let expectVideoID = 'fa5b3be0-7209-4461-8315-17a04c64f5b4';
    httpMock.expectOne(ApiDef.contented.splash).flush(splash);
    fixture.detectChanges();
    tick(10000);
    let url = ApiDef.contented.contentScreens.replace('{mcID}', expectVideoID);
    httpMock.expectOne(url).flush(MockData.getScreens());
  }));
});
