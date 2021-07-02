import { Component } from '@angular/core';
import { ApiService } from 'src/app/_service/api.service';

@Component({
  selector: 'app-root',
  templateUrl: './app.component.html',
  styleUrls: [ './app.component.css' ]
})
export class AppComponent  {
  name = 'Angular';
  title = 'Frontend';

  public data = [];
  public noData: any;
  public results = [];

  constructor(
  private api: ApiService
  ){ }

  getAll() {
    this.api.getAll().subscribe((results) =>  {
      this.data = results.results;
      console.log('JSON Response = ', JSON.stringify(results));
    })
  }

 ngOnInit() {

  }
}
