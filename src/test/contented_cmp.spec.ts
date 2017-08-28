import {async, fakeAsync, getTestBed, tick, ComponentFixture, TestBed} from '@angular/core/testing';
import {By} from '@angular/platform-browser';
import {BaseRequestOptions, Http, Response, ResponseOptions} from '@angular/http';
import {MockBackend, MockConnection} from '@angular/http/testing';
import {DebugElement}    from '@angular/core';

import { RouterTestingModule } from '@angular/router/testing'
import {ContentedCmp} from '../ts/contented/contented_cmp';
import {ContentedService} from '../ts/contented/contented_service';
import {ContentedModule} from '../ts/contented/contented_module';

import * as _ from 'lodash';

describe('TestingContentedsCmp', () => {
    let fixture: ComponentFixture<ContentedCmp>;
    let service: ContentedService;
    let comp: ContentedCmp;
    let mb: MockBackend;
    let el: HTMLElement;
    let de: DebugElement;

    beforeEach(async( () => { 
        TestBed.configureTestingModule({
            imports: [RouterTestingModule, ContentedModule],
            providers: [
                MockBackend,
                BaseRequestOptions,
                {
                    provide: Http,
                    deps: [MockBackend, BaseRequestOptions],
                    useFactory: (mockBackend, options) => {
                        return new Http(mockBackend, options);
                    }
                },
                ContentedService
            ]
        }).compileComponents();

        mb = TestBed.get(MockBackend);
        service = TestBed.get(ContentedService);

        fixture = TestBed.createComponent(ContentedCmp);
        comp = fixture.componentInstance;

        de = fixture.debugElement.query(By.css('.contented-cmp'));
        el = de.nativeElement;
    }));

    it('Should create a contented component', () => {
        expect(comp).toBeDefined("We should have the Contented comp");
        expect(el).toBeDefined("We should have a top level element");
    });
});

