// The browser platform with a compiler
import { platformBrowserDynamic } from '@angular/platform-browser-dynamic';

// The app module
import { AppModule } from './app/app_module';
// import { enableProdMode} from '@angular/core';
// enableProdMode();

// Compile and launch the module
platformBrowserDynamic().bootstrapModule(AppModule);
