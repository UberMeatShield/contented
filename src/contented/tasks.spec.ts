import { ComponentFixture, TestBed, fakeAsync, tick } from '@angular/core/testing';
import { NoopAnimationsModule } from '@angular/platform-browser/animations';
import { HttpTestingController, provideHttpClientTesting } from '@angular/common/http/testing';
import { By } from '@angular/platform-browser';
import { DebugElement } from '@angular/core';

import { ContentedModule } from '../contented/contented_module';
import { TasksCmp } from './tasks.cmp';
import { MockData } from '../test/mock/mock_data';
import { RouterTestingModule } from '@angular/router/testing';
import { provideHttpClient, withInterceptorsFromDi } from '@angular/common/http';

declare var $;

describe('TasksCmp', () => {
  let component: TasksCmp;
  let fixture: ComponentFixture<TasksCmp>;

  let httpMock: HttpTestingController;
  let el: HTMLElement;
  let de: DebugElement;

  beforeEach(() => {
    TestBed.configureTestingModule({
    declarations: [TasksCmp],
    imports: [NoopAnimationsModule,
        ContentedModule,
        RouterTestingModule.withRoutes([{ path: 'admin_ui/tasks', component: TasksCmp }])],
    providers: [provideHttpClient(withInterceptorsFromDi()), provideHttpClientTesting()]
});
    fixture = TestBed.createComponent(TasksCmp);
    component = fixture.componentInstance;

    httpMock = TestBed.inject(HttpTestingController);
    de = fixture.debugElement.query(By.css('.tasks-cmp'));
    el = de.nativeElement;
  });

  afterEach(() => {
    httpMock.verify();
  });

  it('Should be trying to load tasks', fakeAsync(() => {
    fixture.detectChanges();
    let req = httpMock.expectOne(r => r.url.includes('/task_requests'));
    req.flush(MockData.taskRequests());
    tick(1000);
  }));
});
