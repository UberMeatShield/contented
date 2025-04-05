import { fakeAsync, tick, ComponentFixture, TestBed, waitForAsync } from '@angular/core/testing';
import { By } from '@angular/platform-browser';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { DebugElement } from '@angular/core';
import { FormsModule } from '@angular/forms';

import { RouterTestingModule } from '@angular/router/testing';
import { Router } from '@angular/router';

import { TagsCmp } from '../contented/tags.cmp';
import { ContentedService } from '../contented/contented_service';
import { ContentedModule } from '../contented/contented_module';
import { ApiDef } from '../contented/api_def';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';

import * as _ from 'lodash';
import { MockData } from '../test/mock/mock_data';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

describe('Testing TagsCmp', () => {
  let fixture: ComponentFixture<TagsCmp>;
  let service: ContentedService;
  let comp: TagsCmp;
  let el: HTMLElement;
  let de: DebugElement;
  let router: Router;

  let httpMock: HttpTestingController;

  beforeEach(waitForAsync(() => {
    TestBed.configureTestingModule({
    imports: [RouterTestingModule.withRoutes([{ path: 'ui/view/:id', component: TagsCmp }]),
        FormsModule,
        ContentedModule,
        NoopAnimationsModule],
    providers: [ContentedService, provideHttpClient(withInterceptorsFromDi()), provideHttpClientTesting()]
}).compileComponents();

    service = TestBed.inject(ContentedService);
    fixture = TestBed.createComponent(TagsCmp);
    httpMock = TestBed.inject(HttpTestingController);
    comp = fixture.componentInstance;

    de = fixture.debugElement.query(By.css('.tags-cmp'));
    el = de.nativeElement;
    router = TestBed.get(Router);
    router.initialNavigation();
  }));

  afterEach(() => {
    httpMock.verify();
  });

  it('Should create a contented component', () => {
    expect(comp).withContext('We should have the TagsCmp').toBeDefined();
    expect(el).withContext('We should have a top level element').toBeDefined();
  });

  it('Should be able to render or handle the tags in some way', fakeAsync(() => {
    fixture.detectChanges();

    // Could setup search text but really I need a better load monaco language
    // setup for unit tests. Potentially a single environment setup where it loaded
    // up M$ Monaco
    const reqs = httpMock.match(r => r.url.includes('/tags/'));
    for (const req of reqs) {
      req.flush(MockData.tags());
    }
    fixture.detectChanges();
    tick(10000);
  }));
});
