using System.Threading.Tasks;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.DependencyInjection;
using LitterRobot.Models.Shared;
using TwoMQTT.Core;
using LitterRobot.Managers;
using TwoMQTT.Core.Extensions;
using Microsoft.Extensions.Caching.Memory;
using LitterRobot.DataAccess;
using TwoMQTT.Core.DataAccess;

namespace LitterRobot
{
    class Program : ConsoleProgram
    {
        static async Task Main(string[] args)
        {
            var p = new Program();
            await p.ExecuteAsync(args);
        }

        protected override IServiceCollection ConfigureServices(HostBuilderContext hostContext, IServiceCollection services)
        {
            var sharedSect = hostContext.Configuration.GetSection(Models.Shared.Opts.Section);
            var sourceSect = hostContext.Configuration.GetSection(Models.SourceManager.Opts.Section);
            var sinkSect = hostContext.Configuration.GetSection(Models.SinkManager.Opts.Section);

            services.AddHttpClient<SourceManager>();

            services.AddMemoryCache();
            services.AddHttpClient<IHTTPSourceDAO<SlugMapping, Command, Models.SourceManager.FetchResponse, object>>();
            services.AddTransient<IHTTPSourceDAO<SlugMapping, Command, Models.SourceManager.FetchResponse, object>, SourceDAO>();

            return services
                .Configure<Models.Shared.Opts>(sharedSect)
                .Configure<Models.SourceManager.Opts>(sourceSect)
                .Configure<Models.SinkManager.Opts>(sinkSect)
                .ConfigureBidirectionalSourceSink<Resource, Command, SourceManager, SinkManager>();
        }
    }
}