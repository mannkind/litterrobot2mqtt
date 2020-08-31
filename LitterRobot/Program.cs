using System.Collections.Generic;
using System.Net.Http;
using System.Threading.Tasks;
using Microsoft.Extensions.Caching.Memory;
using Microsoft.Extensions.DependencyInjection;
using Microsoft.Extensions.Hosting;
using Microsoft.Extensions.Logging;
using Microsoft.Extensions.Options;
using LitterRobot.DataAccess;
using LitterRobot.Liasons;
using LitterRobot.Models.Shared;
using TwoMQTT.Core;
using TwoMQTT.Core.Extensions;
using TwoMQTT.Core.Interfaces;
using TwoMQTT.Core.Managers;
using System;

namespace LitterRobot
{
    class Program : ConsoleProgram<Resource, Command, SourceLiason, MQTTLiason>
    {
        static async Task Main(string[] args)
        {
            var p = new Program();
            await p.ExecuteAsync(args);
        }

        /// <inheritdoc />
        protected override IDictionary<string, string> EnvironmentDefaults()
        {
            var sep = "__";
            var section = Models.Options.MQTTOpts.Section.Replace(":", sep);
            var sectsep = $"{section}{sep}";

            return new Dictionary<string, string>
            {
                { $"{sectsep}{nameof(Models.Options.MQTTOpts.TopicPrefix)}", Models.Options.MQTTOpts.TopicPrefixDefault },
                { $"{sectsep}{nameof(Models.Options.MQTTOpts.DiscoveryName)}", Models.Options.MQTTOpts.DiscoveryNameDefault },
            };
        }

        /// <inheritdoc />
        protected override IServiceCollection ConfigureServices(HostBuilderContext hostContext, IServiceCollection services)
        {
            services.AddHttpClient<ISourceDAO>();

            return services
                .AddMemoryCache()
                .ConfigureOpts<Models.Options.SharedOpts>(hostContext, Models.Options.SharedOpts.Section)
                .ConfigureOpts<Models.Options.SourceOpts>(hostContext, Models.Options.SourceOpts.Section)
                .ConfigureOpts<TwoMQTT.Core.Models.MQTTManagerOptions>(hostContext, Models.Options.MQTTOpts.Section)
                .AddSingleton<IThrottleManager, ThrottleManager>(x =>
                {
                    var opts = x.GetService<IOptions<Models.Options.SourceOpts>>();
                    if (opts == null)
                    {
                        throw new ArgumentException($"{nameof(opts.Value.PollingInterval)} is required for {nameof(ThrottleManager)}.");
                    }

                    return new ThrottleManager(opts.Value.PollingInterval);
                })
                .AddSingleton<ISourceDAO>(x =>
                {
                    var logger = x.GetService<ILogger<SourceDAO>>();
                    var httpClientFactory = x.GetService<IHttpClientFactory>();
                    var cache = x.GetService<IMemoryCache>();
                    var opts = x.GetService<IOptions<Models.Options.SourceOpts>>();

                    if (logger == null)
                    {
                        throw new ArgumentException($"{nameof(logger)} is required for {nameof(SourceDAO)}.");
                    }
                    if (httpClientFactory == null)
                    {
                        throw new ArgumentException($"{nameof(httpClientFactory)} is required for {nameof(SourceDAO)}.");
                    }
                    if (cache == null)
                    {
                        throw new ArgumentException($"{nameof(cache)} is required for {nameof(SourceDAO)}.");
                    }
                    if (opts == null)
                    {
                        throw new ArgumentException($"{nameof(opts.Value.Login)} and {nameof(opts.Value.Password)} are required for {nameof(SourceDAO)}.");
                    }

                    return new SourceDAO(logger, httpClientFactory, cache, opts.Value.Login, opts.Value.Password
                    );
                });
        }
    }
}